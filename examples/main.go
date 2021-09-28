package main

import "github.com/iamdavidzeng/gonameko"

func main() {
	server := gonameko.Server{
		Name: "gonameko",
	}

	server.Run()
}
