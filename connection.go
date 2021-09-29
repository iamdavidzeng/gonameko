package gonameko

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

type Connection struct {
	Name           string
	RabbitHostname string
	RabbitUser     string
	RabbitPass     string
	RabbitPort     int64
	ContentType    string

	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
	server  amqp.Queue
	msgs    <-chan amqp.Delivery
}

// RPCError capture exception from nameko service
type RPCError struct {
	ExcArgs string `json:"exc_args,omitempty"`
	ExcPath string `json:"exc_path,omitempty"`
	ExcType string `json:"exc_type,omitempty"`
	Value   string `json:"value,omitempty"`
}

// Error represent gonameko customize error
type CustomError struct {
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
	Result interface{} `json:"result"`
	Err    RPCError    `json:"error,omitempty"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("%v: %v", e.ExcType, e.Value)
}

func (e *CustomError) Error() string {
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

	q, err := ch.QueueDeclare(
		fmt.Sprintf("rpc.reply-%v-%v", c.Name, uuid.NewV4().String()), // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	FailOnError(err, "Failed to declare a client queue")
	c.queue = q

	err = ch.QueueBind(
		q.Name,       // name
		q.Name,       // routing key
		"nameko-rpc", // exchange name
		false,        // no-wait
		nil,          // args
	)
	FailOnError(err, "Failed to bind a client queue")

	msgs, err := ch.Consume(
		c.queue.Name, // queue
		"",           // consumer
		false,        // auto ack
		true,         // exclusive
		false,        // no local
		false,        // no wait
		nil,          // args
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
				ReplyTo:       c.queue.Name,
				Body:          []byte(string(param)),
			})
		FailOnError(err, "Failed to publish a message")
	}()

	d := <-c.msgs
	if d.CorrelationId == correlationID {
		json.Unmarshal(d.Body, response)
		serializedResponse, _ := json.Marshal(response)

		log.Println("Got remote response:", string(serializedResponse))

		if (response.Err != RPCError{}) {
			d.Ack(false)
			return nil, &RPCError{response.Err.ExcPath, response.Err.ExcArgs, response.Err.ExcType, response.Err.Value}
		}

		d.Ack(false)
		return response.Result, nil
	}
	return nil, &CustomError{"INVALID_CORRELATION_ID", "invalid correlation id"}

}

func (c *Connection) Serve(s interface{}) {
	instance := s.(*BaseService)
	server, err := c.channel.QueueDeclare(
		fmt.Sprintf("rpc-%v", instance.GetName()), // queue name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
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
		fmt.Sprintf("rpc-%v", instance.GetName()), // queue name
		fmt.Sprintf("%v.*", instance.GetName()),   // routing key
		"nameko-rpc",                              // exchange
		false,                                     // no-wait
		nil,                                       // args
	)
	FailOnError(err, "Failed to bind a server queue")

	msgs, err := c.channel.Consume(
		c.server.Name, // queue name
		"",            // consumer
		false,         // auto ack
		true,          // exclusive
		false,         // no local
		false,         // no wait
		nil,           // args
	)
	FailOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for msg := range msgs {
			log.Println("Server got rpc message: ", msg)

			methodName := ToMethodName(strings.Split(msg.RoutingKey, ".")[1])
			res := reflect.ValueOf(instance).MethodByName(methodName).Call([]reflect.Value{})

			var rpcError RPCError
			val := res[0]
			if len(res) > 1 && res[1].Interface() != nil {
				rpcError = res[1].Interface().(RPCError)
			}
			response, _ := json.Marshal(
				RPCResponse{
					Result: val.Interface(),
					Err:    rpcError,
				},
			)

			log.Println("Server response to", msg.ReplyTo, string(response))
			err = c.channel.Publish(
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
