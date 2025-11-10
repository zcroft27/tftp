package server

import (
	"net"
)

type Server struct {
	port int
	root string
	conn *net.UDPConn
}

func New(port int, root string) *Server {
	return &Server{port: port, root: root}
}

func (s *Server) ListenAndServe() error {
	// Listen for UDP packets
	// Handle RRQ/WRQ requests
	// Spawn goroutines for each transfer
	return nil
}
