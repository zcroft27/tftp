package server

import (
	"context"
	"fmt"
	"log"
	"net"
	protocol "tftp/internal/protocol/parse"
	tftp "tftp/internal/protocol/parse"
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
	localAddr := fmt.Sprintf("localhost:%d", s.port)
	laddr, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		log.Print(err)
		log.Fatal("failed to resolve UDP addr")
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Print(err)
		log.Fatal("failed to create UDP socket:")
	}
	defer conn.Close()
	ctx := context.Background()

	var buf [512]byte
	for {
		n, remote, err := conn.ReadFromUDP(buf[:])
		if err != nil {
			log.Fatalf("failed to read from UDP conn: %w\n", err)
		}
		packet, err := protocol.Parse(buf[:n])
		if err != nil {
			log.Fatalf("failed to parase packet %s\n", string(buf[:n]))
		}
		go handlePacket(ctx, conn, remote, packet)
	}

	return nil
}

func handlePacket(ctx context.Context, conn *net.UDPConn, remote *net.UDPAddr, packet tftp.Packet) {
	fmt.Println("received!")
	fmt.Println(packet.OpCode())
}
