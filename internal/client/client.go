package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	protocol "tftp/internal/protocol/parse"
	"tftp/internal/utils"
	"time"

	humanize "github.com/dustin/go-humanize"
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
	result := make(chan error, 1)

	go get(ctx, result, c.serverAddr, requestingTID, remote, local)

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation timed out: %w", ctx.Err())
	}
}

func (c *Client) Put(local, remote string) error {
	requestingTID := utils.GenerateTID()
	ctx := context.Background()
	result := make(chan error, 1)
	go put(ctx, result, c.serverAddr, requestingTID, remote, local)

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation timed out: %w", ctx.Err())
	}

	return nil
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

func put(ctx context.Context, result chan error, serverAddr string, requestingTID int, remotePath, localPath string) {
	conn, raddr := makeConn(ctx, result, serverAddr, requestingTID)
	if conn == nil || raddr == nil {
		result <- errors.New("failed to make UDP connection")
		return
	}
	defer conn.Close()

	expectedBlockNum := uint16(1)

	// Wait for special-case ACK with block-number 0.

	maxRetries := 5
	timeout := 5 * time.Second
	var serverTIDAddr *net.UDPAddr
	var prevWrq protocol.WriteRequest
	var prevData protocol.Data
	isFirstRequest := true
	offset := int64(0)

	for {
		retries := 0
		buf := make([]byte, TFTP_MAX_DATAGRAM_LENGTH)
		file, err := os.Open(localPath)
		if err != nil {
			result <- err
			return
		}

		defer file.Close()
		n, err := file.ReadAt(buf, offset)
		if err != nil {
			if err != io.EOF {
				result <- err
				return
			}
		}

		for retries < maxRetries {
			if isFirstRequest {
				// Send the WRQ.
				wrq := protocol.WriteRequest{Filename: remotePath, Mode: "netascii"}
				_, err := conn.WriteToUDP(wrq.ToBinary(), raddr)
				if err != nil {
					retries++
					continue
				}
				prevWrq = wrq
			} else {
				// Business as usual.
				packet := protocol.Data{BlockNumber: expectedBlockNum, Data: buf[:n]}
				_, err := conn.WriteToUDP(packet.ToBinary(), serverTIDAddr)
				if err != nil {
					retries++
					continue
				}
			}

			// Wait for ACK.
			buffer := make([]byte, TFTP_MAX_DATAGRAM_LENGTH)
			conn.SetReadDeadline(time.Now().Add(timeout))
			n, addr, err := conn.ReadFromUDP(buffer)

			if err != nil {
				retries++
				// If we've already received ACK, resend the last DATA.
				if expectedBlockNum > 1 {
					conn.WriteToUDP(prevData.ToBinary(), serverTIDAddr)
				}
				continue
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

			if packet.OpCode() != protocol.ACK {
				retries++
				continue
			}

			ack, ok := packet.(protocol.Ack)
			if !ok {
				retries++
				continue
			}

			// Handle retry if the server didn't receive our last DATA/WRQ.
			if ack.BlockNumber != expectedBlockNum {
				if ack.BlockNumber == expectedBlockNum-1 {
					var packetData []byte
					if expectedBlockNum-1 == 0 {
						packetData = prevWrq.ToBinary()
					} else {
						packetData = prevData.ToBinary()
					}
					conn.WriteToUDP(packetData, serverTIDAddr)
				}
				continue
			}

			if expectedBlockNum == 0 {
				isFirstRequest = false
			}

			if serverTIDAddr == nil {
				serverTIDAddr = addr
			}

			break
		}

		if retries >= maxRetries {
			result <- fmt.Errorf("max retries reached for block %d", expectedBlockNum)
		}

		offset += int64(n)
		expectedBlockNum += 1
		if n < TFTP_MAX_DATAGRAM_LENGTH {
			// Transfer complete.
			break
		}
	}

	fmt.Println("Transfer complete")

	result <- nil
}

func get(ctx context.Context, result chan error, serverAddr string, requestingTID int, remotePath, localPath string) {
	conn, raddr := makeConn(ctx, result, serverAddr, requestingTID)
	if conn == nil || raddr == nil {
		result <- errors.New("failed to make UDP connection")
		return
	}
	defer conn.Close()

	file, err := os.Create(localPath)
	if err != nil {
		result <- err
		return
	}
	defer file.Close()

	rrq := protocol.ReadRequest{Filename: remotePath, Mode: "netascii"}
	_, err = conn.WriteToUDP(rrq.ToBinary(), raddr)
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

		n, err := file.Write(dataPacket.Data)
		if n != len(dataPacket.Data) {
			result <- fmt.Errorf("wrote incomplete data into file, aborting")
			return
		}
		if err != nil {
			retries++
			continue
		}

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
	fmt.Printf("Transfer complete: received %s\n", humanize.Bytes(uint64(len(fileData))))

	result <- nil
}
