package tftp

import "encoding/binary"

type OpCode uint16

const (
	RRQ           OpCode = 1 // Read request
	WRQ           OpCode = 2 // Write request
	DATA          OpCode = 3 // Data
	ACK           OpCode = 4 // Acknowledgment
	ERROR         OpCode = 5 // Error
	MODE_NETASCII        = "netascii"
	MODE_OCTET           = "octet"
	MODE_MAIL            = "mail"
)

/*
The  TFTP header consists of a **2 byte** opcode field which indicates

	the packet's type (e.g., DATA, ERROR, etc.)
*/

var VALID_MODES = map[string]bool{
	MODE_NETASCII: true,
	MODE_OCTET:    true,
	MODE_MAIL:     true,
}

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

func (r ReadRequest) ToBinary() []byte {
	opCode := r.OpCode()
	buffer := make([]byte, 2)
	binary.BigEndian.PutUint16(buffer, uint16(opCode))

	filenameBytes := []byte(r.Filename)

	modeBytes := []byte(r.Mode)

	readRequest := []byte{}
	readRequest = append(readRequest, buffer...)
	readRequest = append(readRequest, filenameBytes...)
	readRequest = append(readRequest, 0x00)
	readRequest = append(readRequest, modeBytes...)
	readRequest = append(readRequest, 0x00)

	return readRequest
}

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

func (r WriteRequest) ToBinary() []byte {
	opCode := r.OpCode()
	buffer := make([]byte, 2)
	binary.BigEndian.PutUint16(buffer, uint16(opCode))

	filenameBytes := []byte(r.Filename)

	modeBytes := []byte(r.Mode)

	readRequest := []byte{}
	readRequest = append(readRequest, buffer...)
	readRequest = append(readRequest, filenameBytes...)
	readRequest = append(readRequest, 0x00)
	readRequest = append(readRequest, modeBytes...)
	readRequest = append(readRequest, 0x00)

	return readRequest
}

// Data packet.
type Data struct {
	BlockNumber uint16
	Data        []byte
}

func (d Data) OpCode() OpCode { return DATA }

func (d Data) ToBinary() []byte {
	opCode := d.OpCode()
	opCodeBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(opCodeBuffer, uint16(opCode))

	blockNumberBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(blockNumberBuffer, d.BlockNumber)

	dataPacket := []byte{}
	dataPacket = append(dataPacket, opCodeBuffer...)
	dataPacket = append(dataPacket, blockNumberBuffer...)
	dataPacket = append(dataPacket, d.Data...)

	return dataPacket
}

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
