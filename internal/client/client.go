package client

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"time"
)

type Client struct {
	serverAddr string
}

const (
	TFTP_MAX_DATAGRAM_LENGTH = 516
)

func New(serverAddr string) *Client {
	return &Client{serverAddr: serverAddr}
}

func (c *Client) Get(remote, local string) error {
	TID := 49152 + rand.IntN(65536-49152) // [49152, 65535] is suggested in RFC 6335 as ephemeral ports for dynamic assignment.
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	result := make(chan error, 1)

	go func() {
		get(ctx, result, c.serverAddr, TID, remote, local)
	}()

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation timed out: %w", ctx.Err())
	}
}

func get(ctx context.Context, result chan error, serverAddr string, TID int, remotePath, localPath string) {
	raddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		result <- errors.New("failed to resolve remote UDP address")
		return
	}

	localAddr, err := net.ResolveUDPAddr("udp", "localhost:69")
	if err != nil {
		result <- errors.New("Failed to resolve local UDP address")
		return
	}

	// We must use ListenUDP and not DialUDP since DialUDP creates a 'connected'
	// socket analogous to the connect(2) syscall.
	// However, TFTP requires 1) send on port 69, and 2) continue on port TID,
	// but a connected socket is a socket where the remote address is bound to the socket itself.
	// Therefore, switching ports wouldn't work.
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return
	}

	defer conn.Close()

	// doneChan := make(chan error, 1)

	// buffer := []byte{}
	// go func() {
	// 	n, err := io.Copy(conn, buffer)
	// }()

	deadline, ok := ctx.Deadline()
	if ok {
		conn.SetDeadline(deadline)
	}

	buffer := make([]byte, TFTP_MAX_DATAGRAM_LENGTH)
	n, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		select {
		case <-ctx.Done():
			result <- errors.New("request timed out")
			return
		default:
			result <- fmt.Errorf("failed to receive response: %w", err)
			return
		}
	}

	fmt.Printf("Received %d bytes from %s: %s\n", n, addr, string(buffer[:n]))

	result <- nil
}

func (c *Client) Put(local, remote string) error {
	// Send WRQ
	// Send DATA packets
	// Receive ACKs
	return nil
}
