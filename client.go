package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

type Param struct {
	Args   []string          `json:"args"`
	Kwargs map[string]string `json:"kwargs"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

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
		"go-nameko", // name
		false,       // durable
		false,       // delete when unused
		true,        // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,       // queue name
		q.Name,       // routing key
		"nameko-rpc", // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	corrID := "go_nameko_test_correlation_id"
	param, _ := json.Marshal(Param{
		Args:   []string{},
		Kwargs: map[string]string{},
	})

	err = ch.Publish(
		"nameko-rpc",            // exchange
		"payments.health_check", // routing key
		false,                   // mandatory
		false,                   // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       q.Name,
			Body:          []byte(string(param)),
		})
	failOnError(err, "Failed to publish a message")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			if corrID == d.CorrelationId {
				log.Printf(" [x] %s", d.Body)
			}
		}
	}()
	<-forever
}
