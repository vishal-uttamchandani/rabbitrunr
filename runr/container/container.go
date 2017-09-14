package container

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	units "github.com/docker/go-units"
)

var containers = make([]string, 50)

var (
	// ErrNetworkAlreadyExists is an error that won't cause panic if it were to be raised
	ErrNetworkAlreadyExists = errors.New("Network already exists")
)

const (
	containerNamePreix    = "runr"
	networkName           = "rabbitrunr"
	defaultRabbitMQMemory = "500MB"
	defaultMessageSize    = "1KB"
	maxMessageSize        = "1MB"
	rabbitMQImage         = "rabbitmq:3-management-alpine"
	publisherImage        = "vuttamchandani/publisher"
	consumerImage         = "vuttamchandani/consumer"
)

var (
	c                 *client.Client
	ctx               context.Context
	rbmqContainerName = fmt.Sprintf("%s-rbmq", containerNamePreix)
)

func init() {
	var err error
	c, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	ctx = context.Background()
}

// CreateNetwork will create a network to which all other containers will be connected to
func CreateNetwork() (string, error) {
	networks, err := c.NetworkList(ctx, types.NetworkListOptions{})

	if err != nil {
		return "", err
	}

	for _, n := range networks {
		if n.Name == networkName {
			return "", ErrNetworkAlreadyExists
		}
	}

	nw, err := c.NetworkCreate(ctx, networkName, types.NetworkCreate{})
	if err != nil {
		return "", err
	}

	return nw.ID, nil
}

// CreateRabbitMQ will create a container running rabbitmq borker
func CreateRabbitMQ() (string, error) {

	// key = container port
	// value = host port
	portBindings := map[string]string{
		"15672": "15612",
		"5672":  "5612",
	}

	memory, err := units.FromHumanSize(os.Getenv("rbmq-memory"))
	if err != nil {
		memory, _ = units.FromHumanSize(defaultRabbitMQMemory)
	}

	fmt.Println("memory =", memory)

	return createContainer(rabbitMQImage, rbmqContainerName, nil, nil, nil, portBindings, memory)
}

// CreatePublisher will create a container running a publisher instance
func CreatePublisher(name string, msgSize string) (string, error) {
	links := []string{
		fmt.Sprintf("%s:rbmq", rbmqContainerName),
	}

	envs := []string{
		fmt.Sprintf("qname=%s", name),
		fmt.Sprintf("msg-interval=%d", random(1, 5)),
		fmt.Sprintf("msg-size=%v", getMessageSize(msgSize)),
	}

	containerName := fmt.Sprintf("%s-publisher-%s", containerNamePreix, name)

	return createContainer(publisherImage, containerName, nil, links, envs, nil, 0)
}

func getMessageSize(msgSize string) int64 {
	size, err := units.FromHumanSize(msgSize)
	if err != nil {
		size, _ = units.FromHumanSize(defaultMessageSize)
		fmt.Printf("Unable to infer message size from %s. Using default message size of %s\n", msgSize, defaultMessageSize)
	} else {
		// Max message size supported is 1MB
		if size > 1e6 {
			size = 1e6
			fmt.Printf("Required message size = %s. Max message size supported = %s. Reverting to using message size of %s\n", msgSize, maxMessageSize, maxMessageSize)
		}
	}
	return size
}

// CreateConsumer will create a container running a consumer instance
func CreateConsumer(name string) (string, error) {
	links := []string{
		fmt.Sprintf("%s:rbmq", rbmqContainerName),
	}

	envs := []string{fmt.Sprintf("qname=%s", name)}

	containerName := fmt.Sprintf("%s-consumer-%s", containerNamePreix, name)

	return createContainer(consumerImage, containerName, nil, links, envs, nil, 0)
}

func createContainer(image string, hostname string, labels map[string]string, links []string, envs []string, portBindings map[string]string, memory int64) (string, error) {

	// create port bindings
	portMap := make(nat.PortMap)
	for containerPort, hostPort := range portBindings {
		p, _ := nat.NewPort("tcp", containerPort)
		portMap[p] = []nat.PortBinding{nat.PortBinding{HostIP: "0.0.0.0", HostPort: hostPort}}
	}

	hostConfig := &container.HostConfig{PortBindings: portMap}
	if memory > 0 {
		hostConfig.Memory = memory
	}

	containerConfig := &container.Config{
		Image:    image,
		Hostname: hostname,
		Env:      envs,
		Labels:   labels,
	}

	resp, err := c.ContainerCreate(ctx, containerConfig, hostConfig, nil, hostname)

	if err != nil {
		return "", err
	}

	config := &network.EndpointSettings{
		NetworkID: networkName,
		Links:     links,
	}

	err = c.NetworkConnect(ctx, networkName, resp.ID, config)

	containers = append(containers, resp.ID)

	return resp.ID, nil
}

// Start will start a container
func Start(id string) error {
	if err := c.ContainerStart(ctx, id, types.ContainerStartOptions{}); err != nil {
		return err
	}
	return nil
}

// Stop will stop a container
func Stop(id string) error {
	timeout := time.Duration(1) * time.Minute
	if err := c.ContainerStop(ctx, id, &timeout); err != nil {
		return err
	}
	return nil
}

// Destroy stops and removes all containers created by the runr
func Destroy() string {
	var couldNotDestroy []string
	for _, c := range containers {
		if err := Stop(c); err != nil {
			couldNotDestroy = append(couldNotDestroy, c)
		}

		if err := remove(c); err != nil {
			couldNotDestroy = append(couldNotDestroy, c)
		}
	}

	var totalContainers = len(containers)

	// clear the existing container ids
	containers = containers[:0]

	return fmt.Sprintf("Total containers = %d. Total removed = %d", totalContainers, totalContainers-len(couldNotDestroy))
}

func remove(id string) error {
	if err := c.ContainerRemove(ctx, id, types.ContainerRemoveOptions{}); err != nil {
		return err
	}
	return nil
}

func random(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}
