package gonameko

import (
	"encoding/json"
	"fmt"
	"log"

	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

type Connection struct {
	RabbitHostname string
	RabbitUser     string
	RabbitPass     string
	RabbitPort     int64
	ContentType    string

	conn    *amqp.Connection
	channel *amqp.Channel
	client  amqp.Queue
	server  amqp.Queue
	msgs    <-chan amqp.Delivery
}

// RPCError capture exception from nameko service
type RPCError struct {
	ExcArgs string `json:"exc_args"`
	ExcPath string `json:"exc_path"`
	ExcType string `json:"exc_type"`
	Value   string `json:"value"`
}

// Error represent gonamekoclient customize error
type Error struct {
	Type, Value string
}

// RPCPayload define arguments accept by nameko service
type RPCPayload struct {
	Args   []string          `json:"args"`
	Kwargs map[string]string `json:"kwargs"`
}

// RPCRequestParam define nameko service and function, arguments
type RPCRequestParam struct {
	Service, Function string
	Payload           RPCPayload
}

// RPCResponse Use to parse resposne from nameko service
type RPCResponse struct {
	Result interface{}       `json:"result"`
	Err    map[string]string `json:"error"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("%v: %v", e.ExcType, e.Value)
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v: %v", e.Type, e.Value)
}

func (c *Connection) Declare() {
	amqpURI := fmt.Sprintf(
		"amqp://%v:%v@%v:%v/",
		c.RabbitUser,
		c.RabbitPass,
		c.RabbitHostname,
		c.RabbitPort,
	)
	conn, err := amqp.Dial(amqpURI)
	FailOnError(err, "Failed to connect to RabbitMQ")
	c.conn = conn

	ch, err := conn.Channel()
	FailOnError(err, "Failed to open a channel")
	c.channel = ch

	err = ch.ExchangeDeclare(
		"nameko-rpc", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	FailOnError(err, "Failed to declare an exchange")

	client, err := ch.QueueDeclare(
		"gonameko-client", // name
		false,             // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	FailOnError(err, "Failed to declare a client queue")
	c.client = client

	err = ch.QueueBind(
		client.Name,  // name
		client.Name,  // routing key
		"nameko-rpc", // exchange name
		false,        // no-wait
		nil,          // args
	)
	FailOnError(err, "Failed to bind a client queue")

	msgs, err := ch.Consume(
		c.client.Name, // queue
		"",            // consumer
		false,         // auto ack
		false,         // exclusive
		false,         // no local
		false,         // no wait
		nil,           // args
	)
	FailOnError(err, "Failed to register a consumer")
	c.msgs = msgs
}

func (c *Connection) Call(p RPCRequestParam) (interface{}, error) {
	response := &RPCResponse{}
	correlationID := uuid.NewV4().String()

	go func() {
		param, _ := json.Marshal(p.Payload)

		err := c.channel.Publish(
			"nameko-rpc", // exchange
			fmt.Sprintf("%v.%v", p.Service, p.Function), // routing key
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				ContentType:   c.ContentType,
				CorrelationId: correlationID,
				ReplyTo:       c.client.Name,
				Body:          []byte(string(param)),
			})
		FailOnError(err, "Failed to publish a message")
	}()

	d := <-c.msgs
	if d.CorrelationId == correlationID {
		json.Unmarshal(d.Body, response)

		log.Println(response)

		if response.Err != nil {
			return nil, &RPCError{response.Err["exc_path"], response.Err["exc_args"], response.Err["exc_type"], response.Err["value"]}
		}

		d.Ack(false)

		return response.Result, nil
	}
	return nil, &Error{"INVALID_CORRELATION_ID", "invalid correlation id"}

}

func (c *Connection) Serve(name string) {
	server, err := c.channel.QueueDeclare(
		fmt.Sprintf("rpc-%v", name), // queue name
		false,                       // durable
		false,                       // delete when unused
		false,                       // exclusive
		false,                       // no-wait
		nil,                         // arguments
	)
	FailOnError(err, "Failed to declare a server queue")
	c.server = server

	err = c.channel.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	FailOnError(err, "Failed to set server QoS")

	err = c.channel.QueueBind(
		name,                      // queue name
		fmt.Sprintf("%v.*", name), // routing key
		"nameko-rpc",              // exchange
		false,                     // no-wait
		nil,                       // args
	)
	FailOnError(err, "Failed to bind a queue")

	msgs, err := c.channel.Consume(
		c.server.Name, // queue name
		"",            // consumer
		false,         // auto ack
		false,         // exclusive
		false,         // no local
		false,         // no wait
		nil,           // args
	)
	FailOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for msg := range msgs {
			fmt.Println(msg)
			response, _ := json.Marshal(
				RPCResponse{
					Result: "hello, nameko!",
					Err:    nil,
				},
			)

			err := c.channel.Publish(
				"nameko-rpc",
				msg.ReplyTo,
				false,
				false,
				amqp.Publishing{
					ContentType:   "application/xjson",
					CorrelationId: msg.CorrelationId,
					Body:          response,
				})
			FailOnError(err, "Failed to publish a message")

			msg.Ack(false)
		}
	}()

	log.Printf(" [*] Server is waiting...")
	<-forever
}
