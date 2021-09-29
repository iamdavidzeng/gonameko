package main

import (
	"github.com/iamdavidzeng/gonameko"
)

func main() {
	// client := gonameko.Client{
	// 	RabbitHostname: "localhost",
	// 	RabbitUser:     "guest",
	// 	RabbitPass:     "guest",
	// 	RabbitPort:     5672,
	// 	ContentType:    "application/json",
	// }
	// client.Setup()

	// response, err := client.Call(gonameko.RPCRequestParam{
	// 	Service:  "locations",
	// 	Function: "health_check",
	// 	Payload: gonameko.RPCPayload{
	// 		Args:   []string{},
	// 		Kwargs: map[string]string{},
	// 	},
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(response)
	// }

	service := &gonameko.BaseService{
		Name: "gonameko",
	}

	server := gonameko.Server{
		Name:           "gonameko",
		RabbitHostname: "localhost",
		RabbitUser:     "guest",
		RabbitPass:     "guest",
		RabbitPort:     5672,
		ContentType:    "application/json",
		Service:        service,
	}
	server.Run()
}
