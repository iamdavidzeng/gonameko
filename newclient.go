package gonamekoclient

import (
	"encoding/json"
	"fmt"
	"log"

	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

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

// Client use to initiate a go nameko client
type Client struct {
	RabbitHostname string
	RabbitUser     string
	RabbitPass     string
	RabbitPort     int64
	ContentType    string
	param          RPCPayload

	conn    amqp.Connection
	channel amqp.Channel
	queue   amqp.Queue
	msgs    <-chan amqp.Delivery
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
	Result map[string]interface{} `json:"result"`
	Err    map[string]string      `json:"error"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("%v: %v", e.ExcType, e.Value)
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v: %v", e.Type, e.Value)
}

// Init connected to RabbitMQ and initiate go nameko client
func (r *Client) Init() {
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
	response := &RPCResponse{}
	correlationID := uuid.NewV4().String()

	go func() {
		param, _ := json.Marshal(p.Payload)

		err := r.channel.Publish(
			"nameko-rpc", // exchange
			fmt.Sprintf("%v.%v", p.Service, p.Function), // routing key
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				ContentType:   r.ContentType,
				CorrelationId: correlationID,
				ReplyTo:       r.queue.Name,
				Body:          []byte(string(param)),
			})
		failOnError(err, "Failed to publish a message")
	}()

	d := <-r.msgs
	if d.CorrelationId == correlationID {
		json.Unmarshal(d.Body, response)

		log.Println(response)

		if response.Err != nil {
			return nil, &RPCError{response.Err["exc_path"], response.Err["exc_args"], response.Err["exc_type"], response.Err["value"]}
		}

		return response.Result, nil
	}
	return nil, &Error{"INVALID_CORRELATION_ID", "invalid correlation id"}
}

// Call publish a message to nameko service and return corresponding response
func (r *Client) Call(p RPCRequestParam) (interface{}, error) {
	response, err := r.publish(p)
	return response, err
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
