package main

import "github.com/iamdavidzeng/gonameko"

func main() {
	server := gonameko.Server{
		Name: "gonameko",
	}

	server.Setup(
		"localhost", "guest", "guest", 5672,
	)
	server.Run()
}
