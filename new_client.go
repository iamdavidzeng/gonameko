package main

import (
	"encoding/json"
	"fmt"
	"log"

	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

type RPCError struct {
	Type  string
	Value string
}

type Client struct {
	RabbitHostname string
	RabbitUser     string
	RabbitPass     string
	RabbitPort     int64
	contentType    string
	param          RPCPayload

	conn    amqp.Connection
	channel amqp.Channel
	queue   amqp.Queue
	msgs    <-chan amqp.Delivery
}

type RPCPayload struct {
	Args   []string          `json:"args"`
	Kwargs map[string]string `json:"kwargs"`
}

type RPCRequestParam struct {
	Service, Function string
	Payload           RPCPayload
}

func (e *RPCError) Error() RPCError {
	return RPCError{
		e.Type,
		e.Value,
	}
}

func (r *Client) init() {
	url := fmt.Sprintf("amqp://%v:%v@%v:%v/", r.RabbitUser, r.RabbitPass, r.RabbitHostname, r.RabbitPort)
	conn, err := amqp.Dial(url)
	failOnError(err, "Failed to connect to RabbitMQ")
	r.conn = *conn

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	r.channel = *ch

	err = ch.ExchangeDeclare(
		"nameko-rpc", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"go-nameko-client", // name
		false,              // durable
		false,              // delete when unused
		true,               // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare a queue")
	r.queue = q

	err = ch.QueueBind(
		q.Name,
		q.Name,
		"nameko-rpc",
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		r.queue.Name, // queue
		"",           // consumer
		true,         // auto ack
		false,        // exclusive
		false,        // no local
		false,        // no wait
		nil,          // args
	)
	failOnError(err, "Failed to register a consumer")
	r.msgs = msgs
}

func (r *Client) publish(p RPCRequestParam) (map[string]interface{}, error) {
	response := map[string]interface{}{}
	correlationID := uuid.NewV4().String()

	go func() {
		param, _ := json.Marshal(p.Payload)

		err := r.channel.Publish(
			"nameko-rpc", // exchange
			fmt.Sprintf("%v.%v", p.Service, p.Function), // routing key
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				ContentType:   r.contentType,
				CorrelationId: correlationID,
				ReplyTo:       r.queue.Name,
				Body:          []byte(string(param)),
			})
		failOnError(err, "Failed to publish a message")
	}()

	d := <-r.msgs
	if d.CorrelationId == correlationID {
		json.Unmarshal(d.Body, &response)

		return response, nil
	}
	return nil, nil
}

func (r *Client) request(p RPCRequestParam) (interface{}, error) {
	response, err := r.publish(p)
	return response, err
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	namekorpc := Client{
		RabbitHostname: "localhost",
		RabbitUser:     "guest",
		RabbitPass:     "guest",
		RabbitPort:     5672,
		contentType:    "application/json",
	}

	namekorpc.init()

	response, _ := namekorpc.request(RPCRequestParam{
		Service:  "articles",
		Function: "health_check",
		Payload: RPCPayload{
			Args:   []string{},
			Kwargs: map[string]string{},
		},
	})

	fmt.Println(response)

}
