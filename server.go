package gonameko

type Server struct {
	Name string
	Conn *Connection
}

func (s *Server) Run() {
	s.Conn.Serve(s.Name)
}
