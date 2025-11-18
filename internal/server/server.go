package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"tftp/internal/client"
	protocol "tftp/internal/protocol/parse"
	tftp "tftp/internal/protocol/parse"
	"tftp/internal/utils"
	"time"
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

	var buf [client.TFTP_MAX_DATAGRAM_LENGTH]byte
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
	switch packet.OpCode() {
	case tftp.RRQ:
		// Send DATA
		go sendData(ctx, conn, remote, packet)
	case tftp.WRQ:
		// SEND ACK
		// go sendAck(ctx, conn, remote, packet)
	default:
		// ACK, DATA, and ERROR
		// should never be sent to the server listening at port 69.
		// Well, ERROR packet can be recevied iff the sender received two
		// responses with different TIDs, and the sender rejected one while maintaining the other.
		// ...on the todo list.
	}
}

func sendData(ctx context.Context, conn *net.UDPConn, remote *net.UDPAddr, packet tftp.Packet) {
	TID := utils.GenerateTID()
	newConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: TID})
	if err != nil {
		log.Printf("failed to open conn: %v", err)
		return
	}
	defer newConn.Close()

	fileData := []byte("hello world")
	blockNum := uint16(1)
	offset := 0
	blockSize := client.TFTP_MAX_DATAGRAM_LENGTH
	maxRetries := 5
	timeout := 5 * time.Second

	for {
		end := offset + blockSize
		if end > len(fileData) {
			end = len(fileData)
		}

		dataPacket := tftp.Data{
			BlockNumber: blockNum,
			Data:        fileData[offset:end],
		}

		retries := 0

		for retries < maxRetries {
			_, err := conn.WriteToUDP(dataPacket.ToBinary(), remote)
			if err != nil {
				log.Printf("failed to write: %v", err)
				return
			}

			var buf [client.TFTP_MAX_DATAGRAM_LENGTH]byte
			conn.SetReadDeadline(time.Now().Add(timeout))
			n, _, err := conn.ReadFromUDP(buf[:])

			if err != nil {
				retries++
				log.Printf("timeout waiting for ACK block %d (attempt %d/%d)",
					blockNum, retries, maxRetries)
				continue
			}

			ackPacket, err := tftp.Parse(buf[:n])
			if err != nil || ackPacket.OpCode() != tftp.ACK {
				log.Printf("invalid packet received, expected ACK")
				retries++
				continue
			}

			ack, ok := ackPacket.(tftp.Ack)
			if !ok || ack.BlockNumber != blockNum {
				log.Printf("wrong ACK block number, expected %d got %d",
					blockNum, ack.BlockNumber)
				retries++
				continue
			}

			// ACK received successfully.
			break
		}

		if retries >= maxRetries {
			log.Printf("max retries reached for block %d, aborting", blockNum)
			return
		}

		// Check if transfer complete (last block < client.TFTP_MAX_DATAGRAM_LENGTH bytes).
		if end-offset < blockSize {
			log.Printf("transfer complete")
			return
		}

		blockNum++
		offset = end
	}
}
