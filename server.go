package gonameko

type Server struct {
	Name string

	RabbitHostname string
	RabbitUser     string
	RabbitPass     string
	RabbitPort     int64
	ContentType    string

	Conn    *Connection
	Service interface{}
}

func (s *Server) Run() {
	s.Conn = &Connection{
		Name:           s.Name,
		RabbitHostname: s.RabbitHostname,
		RabbitUser:     s.RabbitUser,
		RabbitPass:     s.RabbitPass,
		RabbitPort:     s.RabbitPort,
		ContentType:    s.ContentType,
	}
	s.Conn.Declare()
	s.Conn.Serve(s.Service)
}
