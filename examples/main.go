package main

import "github.com/iamdavidzeng/gonameko"

func main() {
	server := gonameko.Server{
		Name:           "gonameko",
		RabbitHostname: "localhost",
		RabbitUser:     "guest",
		RabbitPass:     "guest",
		RabbitPort:     5672,
		ContentType:    "application/json",
	}

	server.Run()
}
