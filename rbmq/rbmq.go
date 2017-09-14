package rbmq

import (
	"fmt"

	"github.com/streadway/amqp"
)

const (
	appName = "rabbitrunr"
)

var q amqp.Queue

// Context represents connection and channel to the broker
type Context struct {
	Conn *amqp.Connection
	Ch   *amqp.Channel
}

var (
	// ErrInvalidQueue is raised either when queue is not initialized on the broker
	ErrInvalidQueue = "Queue does not exists or is invalid. Make sure you call DeclareQueue before calling Publish"
)

// Dial connects to the rabbitmq broker running at the given url
func Dial(url string) (*Context, error) {
	var err error
	var conn *amqp.Connection
	var ch *amqp.Channel

	if conn, err = amqp.Dial(url); err != nil {
		return nil, err
	}

	if ch, err = conn.Channel(); err != nil {
		return nil, err
	}

	return &Context{conn, ch}, nil
}

// DeclareQueue ensures that the queue with provided properties exists in rabbitmq and if it doesn't then a new queue is created
func (r *Context) DeclareQueue(qname string) error {
	var err error
	q, err = r.Ch.QueueDeclare(qname, true, false, false, false, nil)
	if err != nil {
		return err
	}
	return nil
}

// Publish will publish message to the queue
func (r *Context) Publish(msg string) error {
	if q.Name == "" {
		panic(ErrInvalidQueue)
	}

	p := amqp.Publishing{
		AppId:        appName,
		DeliveryMode: amqp.Persistent, // make messages persistent
		ContentType:  "text/plain",
		Body:         []byte(msg),
	}

	if err := r.Ch.Publish("", q.Name, false, false, p); err != nil {
		return err
	}

	return nil
}

// NotifyEvents will present useful messages arising from events happening on the connection and channel
func (r *Context) NotifyEvents(notification chan string) chan string {
	go func() {
		//var blocked bool
		for {
			n := <-r.Conn.NotifyBlocked(make(chan amqp.Blocking))
			//if n.Active != blocked {
			notification <- fmt.Sprintf("Blocked notification received with values - Active: %t, Reason: %s\n", n.Active, n.Reason)
			//blocked = n.Active
			//}
		}
	}()

	go func() {
		n := <-r.Conn.NotifyClose(make(chan *amqp.Error))
		notification <- fmt.Sprintf("Closed notification received on connection with values - Reason: %s, Initiated by server: %t, Recovery possible: %t\n", n.Reason, n.Server, n.Recover)
	}()

	go func() {
		n := <-r.Ch.NotifyCancel(make(chan string))
		notification <- fmt.Sprintln(n)
	}()

	go func() {
		n := <-r.Ch.NotifyClose(make(chan *amqp.Error))
		notification <- fmt.Sprintf("Closed notification received on channel with values - Reason: %s, Initiated by server: %t, Recovery possible: %t\n", n.Reason, n.Server, n.Recover)
	}()

	go func() {
		//var flowControl bool
		for {
			n := <-r.Ch.NotifyFlow(make(chan bool))
			//if n != flowControl {
			notification <- fmt.Sprintf("Flow control = %t", n)
			//flowControl = n
			//}
		}
	}()

	return notification
}

// Consume will consume message from a given queue
func (r *Context) Consume(qname string) (<-chan amqp.Delivery, error) {
	deliveries, err := r.Ch.Consume(
		qname, // queue name
		"",    // consumer tag
		true,  // noAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // arguments
	)

	return deliveries, err
}

// Close will close the connection and channel
func (r *Context) Close() error {
	var err error
	if r.Ch != nil {
		if err = r.Ch.Close(); err != nil {
			return err
		}
	}
	if r.Conn != nil {
		if err = r.Conn.Close(); err != nil {
			return err
		}
	}

	return nil
}
