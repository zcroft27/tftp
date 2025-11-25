package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"tftp/internal/client"
	protocol "tftp/internal/protocol/parse"
	tftp "tftp/internal/protocol/parse"
	"tftp/internal/utils"
	"time"

	"github.com/dustin/go-humanize"
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
		go handlePacket(ctx, remote, packet, s.root)
	}
}

func handlePacket(ctx context.Context, remote *net.UDPAddr, packet tftp.Packet, root string) {
	switch packet.OpCode() {
	case tftp.RRQ:
		rrq, ok := packet.(tftp.ReadRequest)
		if !ok {
			log.Printf("failed to convert to RRQ")
			return
		}
		filename := rrq.Filename
		fullPath := filepath.Join(root, filename)

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			log.Printf("file not found: %s", filename)
			// TODO: send ERROR packet to client.
			return
		}

		fileData, err := os.ReadFile(fullPath)
		if err != nil {
			log.Printf("failed to read file %s: %v", filename, err)
			// TODO: send ERROR packet to client.
			return
		}

		// Send DATA
		handleRRQ(ctx, remote, fileData)
	case tftp.WRQ:
		// SEND ACK
		wrq, ok := packet.(protocol.WriteRequest)
		if !ok {
			log.Printf("failed to convert packet: %w\n", packet)
			return
		}
		go func() {
			fullPath := filepath.Join(root, wrq.Filename)
			file, err := os.Create(fullPath)
			if err != nil {
				fmt.Println("err: ", err)
				log.Printf("failed to open file: %s\n", wrq.Filename)
				// TODO: send ERROR packet.
				return
			}
			defer file.Close()
			handleWRQ(ctx, remote, file)
		}()
	default:
		// ACK, DATA, and ERROR
		// should never be sent to the server listening at port 69.
		// Well, ERROR packet can be recevied iff the sender received two
		// responses with different TIDs, and the sender rejected one while maintaining the other.
		// ...on the todo list.
	}
}

func handleWRQ(ctx context.Context, remote *net.UDPAddr, file *os.File) {
	// TID := utils.GenerateTID()
	log.Printf("Remote address: %v (IP: %v, Port: %d)", remote, remote.IP, remote.Port)

	newConn, err := net.DialUDP("udp", nil, remote)
	// newConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: TID})
	if err != nil {
		log.Printf("failed to open conn: %v", err)
		return
	}
	defer newConn.Close()

	blockNum := uint16(0)
	maxRetries := 5
	timeout := 5 * time.Second
	var fileData []byte
	var dataPacket protocol.Data

	for {
		select {
		case <-ctx.Done():
			log.Printf("context cancelled: %v", ctx.Err())
			return
		default:
			// Continue with transfer.
		}

		retries := 0
		for retries < maxRetries {
			ackPacket := protocol.Ack{BlockNumber: blockNum}
			_, err := newConn.Write(ackPacket.ToBinary())
			if err != nil {
				retries++
				continue
			}

			buffer := make([]byte, 1024)
			newConn.SetReadDeadline(time.Now().Add(timeout))
			n, err := newConn.Read(buffer)

			if err != nil {
				retries++
				continue
			}

			packet, err := protocol.Parse(buffer[:n])
			if err != nil {
				log.Printf("failed to parse packet: %w", err)
				return
			}

			// Check if it's an ERROR packet.
			if packet.OpCode() == protocol.ERROR {
				errorPacket, ok := packet.(protocol.Error)
				if ok {
					log.Printf("server error %d: %s", errorPacket.ErrorCode, errorPacket.ErrorMsg)
				} else {
					log.Println("received error packet from server")
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

			expectedDataBlock := blockNum + 1

			// Handle retry if the server didn't receive our last ACK.
			if data.BlockNumber != expectedDataBlock {
				if data.BlockNumber == blockNum {
					ackPacket := protocol.Ack{BlockNumber: data.BlockNumber}
					newConn.Write(ackPacket.ToBinary())
				}
				continue
			}

			dataPacket = data
			break
		}

		if retries >= maxRetries {
			log.Printf("max retries reached for block %d", blockNum)
			return
		}

		fileData = append(fileData, dataPacket.Data...)

		n, err := file.Write(dataPacket.Data)
		if n != len(dataPacket.Data) {
			log.Println("wrote incomplete data into file, aborting")
			return
		}
		if err != nil {
			retries++
			continue
		}

		blockNum++

		if n < client.TFTP_MAX_DATAGRAM_LENGTH {
			fmt.Printf("Transfer complete: received %s\n", humanize.Bytes(uint64(len(fileData))))
			return
		}

	}
}

func handleRRQ(ctx context.Context, remote *net.UDPAddr, fileData []byte) {
	TID := utils.GenerateTID()
	newConn, err := net.DialUDP("udp", &net.UDPAddr{Port: TID}, remote)
	// newConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: TID})
	if err != nil {
		log.Printf("failed to open conn: %v", err)
		return
	}
	defer newConn.Close()

	blockNum := uint16(1)
	offset := 0
	blockSize := client.TFTP_MAX_DATAGRAM_LENGTH
	maxRetries := 5
	timeout := 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			log.Printf("context cancelled: %v", ctx.Err())
			return
		default:
			// Continue with transfer.
		}

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
			_, err := newConn.Write(dataPacket.ToBinary())
			if err != nil {
				log.Printf("failed to write: %v", err)
				return
			}

			var buf [client.TFTP_MAX_DATAGRAM_LENGTH]byte
			newConn.SetReadDeadline(time.Now().Add(timeout))
			n, err := newConn.Read(buf[:])

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
