package tftp

import "fmt"

/*
The  TFTP header consists of a **2 byte** opcode field which indicates

	the packet's type (e.g., DATA, ERROR, etc.)
*/
type OpCode uint16

const (
	RRQ   OpCode = 1 // Read request
	WRQ   OpCode = 2 // Write request
	DATA  OpCode = 3 // Data
	ACK   OpCode = 4 // Acknowledgment
	ERROR OpCode = 5 // Error
)

// Packet is the interface all TFTP packets implement.
type Packet interface {
	OpCode() OpCode
}

// ReadRequest packet.
/*
The mode field contains the
   string "netascii", "octet", or "mail" (or any combination of upper
   and lower case, such as "NETASCII", NetAscii", etc.) in netascii
   indicating the three modes defined in the protocol.
*/
type ReadRequest struct {
	Filename string
	Mode     string // "netascii", "octet", or "mail".
}

func (r ReadRequest) OpCode() OpCode { return RRQ }

// WriteRequest packet.
/*
The mode field contains the
   string "netascii", "octet", or "mail" (or any combination of upper
   and lower case, such as "NETASCII", NetAscii", etc.) in netascii
   indicating the three modes defined in the protocol.
*/
type WriteRequest struct {
	Filename string
	Mode     string // "netascii", "octet", or "mail".
}

func (w WriteRequest) OpCode() OpCode { return WRQ }

// Data packet.
type Data struct {
	BlockNumber uint16
	Data        []byte
}

func (d Data) OpCode() OpCode { return DATA }

// Acknowledgment packet.
type Ack struct {
	BlockNumber uint16
}

func (a Ack) OpCode() OpCode { return ACK }

// Error packet.
type Error struct {
	ErrorCode uint16
	ErrorMsg  string
}

func (e Error) OpCode() OpCode { return ERROR }

// Parse parses raw bytes into a TFTP packet.
// Chosen to be built on top of UDP, and UDP datagram is 1:1 with TFTP packet.
func Parse(data []byte) (Packet, error) {
	return nil, fmt.Errorf("not implemented")
}
