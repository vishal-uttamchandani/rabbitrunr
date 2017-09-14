package main

import (
	"fmt"
	"log"
	"net/http"
	"rabbitrunr/runr/container"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

type runrParams struct {
	MessageInterval int    `json:"msg-interval"`
	DeviceCount     int    `json:"device-count"`
	MessageSize     string `json:"msg-size"`
}

type containerContext struct {
	RunrType  string `json:"runr-type"`
	QueueName string `json:"queue-name"`
}

var (
	r          *gin.Engine
	containers []string
)

func main() {
	// Start a rabbitmq container
	if err := initialize(); err != nil {
		panic(err)
	}

	// Initialize the gin router
	r = gin.Default()

	// Setup routes and route handlers
	initializeRoutes()

	// Listen on port 5000
	r.Run(":5000")
}

func initializeRoutes() {
	r.GET("/info", handleInfo)
	r.POST("/runr", handleRunr)
	r.POST("/destroy", handleDestroy)
	r.POST("/restart", handleRestart)
	r.POST("/stop", handleStop)
	r.POST("/start", handleStart)
}

func handleStop(c *gin.Context) {
	var err error
	var p containerContext

	c.BindJSON(&p)

	name, err := p.getContainerName()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// stop the container
	if err = container.Stop(name); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("could not stop '%s'", name))
		return
	}

	c.JSON(http.StatusOK, fmt.Sprintf("'%s' stopped successfully", name))
}

func handleStart(c *gin.Context) {
	var p containerContext
	c.BindJSON(&p)

	// get the container name based on the queue-name and runr type
	name, err := p.getContainerName()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// start the container
	if err := container.Start(name); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("could not start '%s'", name))
		return
	}

	c.JSON(http.StatusOK, fmt.Sprintf("'%s' started successfully", name))
}

func handleRestart(c *gin.Context) {
	// Stop and remove all containers created by the runr
	container.Destroy()

	// Start a rabbitmq container
	if err := initialize(); err != nil {
		panic(err)
	}
}

func handleInfo(c *gin.Context) {
	c.JSON(http.StatusOK, "RabbitRunr version 1.0")
}

func handleRunr(c *gin.Context) {
	var p runrParams
	c.BindJSON(&p)

	for i := 0; i < p.DeviceCount; i++ {
		// create publisher, queue and consumer (PQC)
		createPQC(p)
	}

	c.JSON(http.StatusCreated, fmt.Sprintf("%d publishers, queues, and consumers created", p.DeviceCount))
}

func createPQC(p runrParams) {
	qname := uuid.NewV4().String()
	id, err := container.CreatePublisher(qname, p.MessageSize)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("publisher container created with id =", id)
	}

	if err := container.Start(id); err != nil {
		log.Fatal(err)
	}

	log.Println("publisher container started with id =", id)

	id, err = container.CreateConsumer(qname)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("consumer container created with id =", id)
	}
}

func handleDestroy(c *gin.Context) {
	// Stop and remove all containers created by the runr
	msg := container.Destroy()

	c.JSON(http.StatusOK, msg)
}

func initialize() error {
	_, err := container.CreateNetwork()
	if err != nil {
		if err == container.ErrNetworkAlreadyExists {
			fmt.Println("Using existing network")
		} else {
			return err
		}
	}

	id, err := container.CreateRabbitMQ()
	if err != nil {
		return err
	}

	containers = append(containers, id)

	err = container.Start(id)
	if err != nil {
		return err
	}

	return nil
}

func (p containerContext) getContainerName() (string, error) {
	var containerName string
	switch p.RunrType {
	case "publisher":
		containerName = fmt.Sprintf("runr-publisher-%s", p.QueueName)
	case "consumer":
		containerName = fmt.Sprintf("runr-consumer-%s", p.QueueName)
	}

	if containerName == "" {
		return "", fmt.Errorf("Uknown runr type '%s'", p.RunrType)
	}

	return containerName, nil
}
