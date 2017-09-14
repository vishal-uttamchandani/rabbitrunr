package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"rabbitrunr/publisher/data"
	"strconv"
	"syscall"
	"time"

	"github.com/vishal-uttamchandani/rabbitrunr/rbmq"
)

var ctx *rbmq.Context

const (
	defaultMessageSize = "1KB"
)

func main() {
	var err error
	go handleShutdown()

	qname := os.Getenv("qname")
	msgInterval, _ := strconv.Atoi(os.Getenv("msg-interval"))
	msgSize, _ := strconv.ParseInt(os.Getenv("msg-size"), 10, 64)

	ctx, err = rbmq.Dial("amqp://guest:guest@rbmq:5672")
	failOnError(err)

	err = ctx.DeclareQueue(qname)
	failOnError(err)

	fmt.Printf("publisher will publish to '%s' queue every %d seconds\n", qname, msgInterval)

	go handleNotifications()

	for {
		<-time.Tick(time.Duration(msgInterval) * time.Second)
		if err := ctx.Publish(data.Data[:msgSize]); err != nil {
			log.Println(err)
		}
	}
}

func handleNotifications() {
	for {
		n := <-ctx.NotifyEvents(make(chan string, 100))
		fmt.Println(n)
	}
}

func handleShutdown() {
	// signal channel to received notifications
	sigc := make(chan os.Signal, 1)

	// Signals to handle
	signals := []os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	}

	// Register the channel for notification of signals
	signal.Notify(sigc, signals...)

	go func() {
		s := <-sigc
		fmt.Printf("publisher stopped. Received signal = %#v\n", s)
	}()
}

func failOnError(err error) {
	if err != nil {
		panic(err)
	}
}
