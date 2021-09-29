package gonameko

type Server struct {
	Name string
	Conn *Connection
}

func (s *Server) Run() {
	s.Conn.Serve(s.Name)
}

func (s *Server) Setup(host, user, pass string, port int64) {
	s.Conn = &Connection{
		RabbitHostname: host,
		RabbitUser:     user,
		RabbitPass:     pass,
		RabbitPort:     port,
		ContentType:    "application/xjson",
	}
	s.Conn.Declare()
}
