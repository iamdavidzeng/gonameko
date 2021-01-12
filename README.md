# go-nameko
A Golang client of nameko

# Tutorial
```
go get github.com/iamdavidzeng/gonamekoclient
```

## Usage
```
import gonameko "github.com/iamdavidzeng/gonamekoclient"

func main() {
    namekorpc := gonameko.Client{
        RabbitHostname: "localhost",
        RabbitUser: "guest",
        RabbitPass: "guest",
        RabbitPort: 5672,
        ContentType: "application/json",
    }

    namekorpc.Init()

    response, _ = namekorpc.request(
        RPCRequestParam{
            Service: "articles",
            Function: "health_check",
            Payload: RPCPayload{
                Args: []string{},
                Kwargs: map[string]string{},
            },
        },
    )
}
```