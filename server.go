package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type Server struct {
	Name           string
	RabbitHostname string
	RabbitUser     string
	RabbitPass     string
	RabbitPort     int64
	ContentType    string

	conn    *amqp.Connection
	channel *amqp.Channel
	queue   *amqp.Queue
	msgs    <-chan amqp.Delivery
}

type RPCResponse struct {
	Result interface{}       `json:"result"`
	Err    map[string]string `json:"error"`
}

func (s *Server) Run() {
	url := fmt.Sprintf("amqp://%v:%v@%v:%v/", s.RabbitUser, s.RabbitPass, s.RabbitHostname, s.RabbitPort)
	conn, err := amqp.Dial(url)
	FailOnError(err, "Failed to connect to RabbitMQ")
	s.conn = conn

	ch, err := conn.Channel()
	FailOnError(err, "Failed to open a channel")
	s.channel = ch

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
		fmt.Sprintf("rpc-%v", s.Name), // name
		false,                         // durable
		false,                         // delete when unused
		true,                          // exclusive
		false,                         // no-wait
		nil,                           // arguments
	)
	FailOnError(err, "Failed to declare a queue")
	s.queue = &q

	err = ch.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	FailOnError(err, "Failed to set QoS")

	err = ch.QueueBind(
		q.Name,
		fmt.Sprintf("%v.*", s.Name),
		"nameko-rpc",
		false,
		nil,
	)
	FailOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		s.queue.Name, // queue
		"",           // consumer
		true,         // auto ack
		false,        // exclusive
		false,        // no local
		false,        // no wait
		nil,          // args
	)
	FailOnError(err, "Failed to register a consumer")
	s.msgs = msgs
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	server := Server{
		Name:           "gonameko",
		RabbitHostname: "localhost",
		RabbitUser:     "guest",
		RabbitPass:     "guest",
		RabbitPort:     5672,
		ContentType:    "application/json",
	}

	server.Run()

	forever := make(chan bool)

	go func() {
		for msg := range server.msgs {
			fmt.Println(string(msg.Body))
			response, _ := json.Marshal(
				RPCResponse{
					Result: "hello, nameko!",
					Err:    nil,
				},
			)

			err := server.channel.Publish(
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
