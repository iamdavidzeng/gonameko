# go-nameko
A Golang client of nameko

# Tutorial
```
go get -u github.com/iamdavidzeng/gonamekoclient
```

## Usage
```
package main

import (
	"fmt"

	"github.com/iamdavidzeng/gonamekoclient"
)

func main() {
	namekorpc := gonamekoclient.Client{
		RabbitHostname: "localhost",
		RabbitUser:     "guest",
		RabbitPass:     "guest",
		RabbitPort:     5672,
		ContentType:    "application/json",
	}

	namekorpc.Init()

	response, err := namekorpc.Call(
		gonamekoclient.RPCRequestParam{
			Service:  "articles",
			Function: "health_check",
			Payload: gonamekoclient.RPCPayload{
				Args:   []string{},
				Kwargs: map[string]string{},
			},
		},
	)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(response)
	}
}
```
