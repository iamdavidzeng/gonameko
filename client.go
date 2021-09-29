package gonameko

// Client use to initiate a go nameko client
type Client struct {
	Conn *Connection
}

// Call publish a message to nameko service and return corresponding response
func (c *Client) Call(p RPCRequestParam) (interface{}, error) {
	response, err := c.Conn.Call(p)
	return response, err
}

func (c *Client) Setup(host, user, pass string, port int64) {
	c.Conn = &Connection{
		RabbitHostname: host,
		RabbitUser:     user,
		RabbitPass:     pass,
		RabbitPort:     port,
		ContentType:    "application/xjson",
	}
	c.Conn.Declare()
}
