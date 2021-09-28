package gonameko

import "github.com/streadway/amqp"

type Connection struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   *amqp.Queue
}

func (c *Connection) Dial() {}
