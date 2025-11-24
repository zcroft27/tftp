package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	protocol "tftp/internal/protocol/parse"
	"tftp/internal/utils"
	"time"
)

type Client struct {
	serverAddr string
}

const (
	TFTP_MAX_DATAGRAM_LENGTH = 512
)

func New(serverAddr string) *Client {
	return &Client{serverAddr: serverAddr}
}

func (c *Client) Get(remote, local string) error {
	requestingTID := utils.GenerateTID()
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	result := make(chan error, 1)

	go get(ctx, result, c.serverAddr, requestingTID, remote, local)

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation timed out: %w", ctx.Err())
	}
}

func makeConn(ctx context.Context, result chan error, serverAddr string, requestingTID int) (*net.UDPConn, *net.UDPAddr) {
	raddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		result <- errors.New("failed to resolve remote UDP address")
		return nil, nil
	}

	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", requestingTID))
	if err != nil {
		result <- errors.New("failed to resolve local UDP address")
		return nil, nil
	}

	// We must use ListenUDP and not DialUDP since DialUDP creates a 'connected'
	// socket analogous to the connect(2) syscall.
	// However, TFTP requires 1) send on port 69, and 2) continue on port TID,
	// but a connected socket is a socket where the remote address is bound to the socket itself.
	// Therefore, switching ports wouldn't work.
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		result <- err
		return nil, nil
	}

	// defer connection close in caller.

	deadline, ok := ctx.Deadline()
	if ok {
		conn.SetDeadline(deadline)
	}

	return conn, raddr
}

func get(ctx context.Context, result chan error, serverAddr string, requestingTID int, remotePath, localPath string) {
	conn, raddr := makeConn(ctx, result, serverAddr, requestingTID)
	if conn == nil || raddr == nil {
		result <- errors.New("failed to make UDP connection")
		return
	}
	defer conn.Close()

	message := protocol.ReadRequest{Filename: remotePath, Mode: "netascii"}
	_, err := conn.WriteToUDP(message.ToBinary(), raddr)
	if err != nil {
		result <- err
		return
	}

	var fileData []byte
	expectedBlockNum := uint16(1)
	maxRetries := 5
	timeout := 5 * time.Second
	var serverTIDAddr *net.UDPAddr

	for {
		retries := 0
		var dataPacket protocol.Data

		for retries < maxRetries {
			buffer := make([]byte, 1024)
			conn.SetReadDeadline(time.Now().Add(timeout))
			n, addr, err := conn.ReadFromUDP(buffer)

			if err != nil {
				retries++
				// If we've already received data, resend the last ACK.
				if expectedBlockNum > 1 {
					ackPacket := protocol.Ack{BlockNumber: expectedBlockNum - 1}
					conn.WriteToUDP(ackPacket.ToBinary(), serverTIDAddr)
				}
				continue
			}

			if serverTIDAddr == nil {

				serverTIDAddr = addr
			}

			packet, err := protocol.Parse(buffer[:n])
			if err != nil {
				result <- fmt.Errorf("failed to parse packet: %w", err)
				return
			}

			// Check if it's an ERROR packet.
			if packet.OpCode() == protocol.ERROR {
				errorPacket, ok := packet.(protocol.Error)
				if ok {
					result <- fmt.Errorf("server error %d: %s", errorPacket.ErrorCode, errorPacket.ErrorMsg)
				} else {
					result <- errors.New("received error packet from server")
				}
				return
			}

			if packet.OpCode() != protocol.DATA {
				retries++
				continue
			}

			data, ok := packet.(protocol.Data)
			if !ok {
				retries++
				continue
			}

			// Handle retry if the server didn't receive our last ACK.
			if data.BlockNumber != expectedBlockNum {
				if data.BlockNumber == expectedBlockNum-1 {
					ackPacket := protocol.Ack{BlockNumber: data.BlockNumber}
					conn.WriteToUDP(ackPacket.ToBinary(), serverTIDAddr)
				}
				continue
			}

			dataPacket = data
			break
		}

		if retries >= maxRetries {
			result <- fmt.Errorf("max retries reached for block %d", expectedBlockNum)
			return
		}

		fileData = append(fileData, dataPacket.Data...)

		// Send ACK.
		ackPacket := protocol.Ack{BlockNumber: expectedBlockNum}
		_, err = conn.WriteToUDP(ackPacket.ToBinary(), serverTIDAddr)
		if err != nil {
			result <- fmt.Errorf("failed to send ACK: %w", err)
			return
		}

		// Check if this was the last block.
		if len(dataPacket.Data) < TFTP_MAX_DATAGRAM_LENGTH {
			// Transfer complete.
			break
		}

		expectedBlockNum++
	}

	// TODO: Write fileData to localPath.
	fmt.Printf("Transfer complete: received %d bytes\n", len(fileData))
	fmt.Printf("data: %s\n", fileData)
	result <- nil
}

func (c *Client) Put(local, remote string) error {
	// Send WRQ
	// Send DATA packets
	// Receive ACKs
	return nil
}
