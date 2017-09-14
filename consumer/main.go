package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vishal-uttamchandani/rabbitrunr/rbmq"
)

var ctx *rbmq.Context

func main() {
	var err error
	go handleShutdown()

	qname := os.Getenv("qname")

	ctx, err = rbmq.Dial("amqp://guest:guest@rbmq:5672")
	failOnError(err)

	err = ctx.DeclareQueue(qname)
	failOnError(err)

	messages, err := ctx.Consume(qname)
	failOnError(err)

	fmt.Println("Consuming messages..")
	for range messages {
		//fmt.Println(string(msg.Body))
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
		fmt.Printf("consumer stopped. Received signal = %#v\n", s)
	}()
}

func failOnError(err error) {
	if err != nil {
		panic(fmt.Sprintf("%s", err))
	}
}
