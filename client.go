package gonameko

// Client use to initiate a go nameko client
type Client struct {
	RabbitHostname string
	RabbitUser     string
	RabbitPass     string
	RabbitPort     int64
	ContentType    string

	Conn *Connection
}

// Call publish a message to nameko service and return corresponding response
func (c *Client) Call(p RPCRequestParam) (interface{}, error) {
	response, err := c.Conn.Call(p)
	return response, err
}

func (c *Client) Setup() {
	c.Conn = &Connection{
		Name:           "gonameko-client",
		RabbitHostname: c.RabbitHostname,
		RabbitUser:     c.RabbitUser,
		RabbitPass:     c.RabbitPass,
		RabbitPort:     c.RabbitPort,
		ContentType:    c.ContentType,
	}
	c.Conn.Declare()
}
