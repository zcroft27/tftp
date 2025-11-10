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
	TID := 49152 + rand.IntN(65536-49152) // 49152 is suggested in RFC 6335 as ephemeral ports for dynamic assignment.
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
	serverUDPAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		result <- fmt.Errorf("failed to resolve server address: %w", err)
		return
	}

	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", TID))
	if err != nil {
		result <- fmt.Errorf("failed to resolve local address: %w", err)
		return
	}

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		result <- fmt.Errorf("failed to create UDP connection: %w", err)
		return
	}
	defer conn.Close()

	message := []byte("Hello TFTP Server!")

	n, err := conn.WriteToUDP(message, serverUDPAddr)
	if err != nil {
		result <- fmt.Errorf("failed to send UDP message: %w", err)
		return
	}

	fmt.Printf("Sent %d bytes to %s from local port %d\n", n, serverAddr, TID)

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
