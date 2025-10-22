package tftp

import (
	"encoding/binary"
	"errors"
	"fmt"
)

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
	if len(data) <= 0 {
		return nil, errors.New("no data to parse")
	}

	opcode := OpCode(binary.BigEndian.Uint16(data[0:2]))

	parsers := map[OpCode]func([]byte) (Packet, error){
		RRQ:   parseReadRequest,
		WRQ:   parseWriteRequest,
		ACK:   parseAckRequest,
		DATA:  parseDataRequest,
		ERROR: parseErrorRequest,
	}

	parser, exists := parsers[opcode]
	if !exists {
		return nil, errors.New("unrecognized opcode")
	}

	packet, err := parser(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data of opcode type: %v", opcode)
	}

	return packet, errors.New("unimplemented")
}

func parseReadRequest(data []byte) (Packet, error) {
	if len(data) < 4 {
		return nil, errors.New("RRQ packet is missing opcode and/or required delimiters.")
	}

	var packet ReadRequest

	var filenameBytes []byte
	endFilenameIdx := -1 // the zero delimiter idx after a string.
	for idx, by := range data[2:] {
		if by == 0x0 {
			endFilenameIdx = idx
			break
		}
		filenameBytes = append(filenameBytes, by)
	}

	if endFilenameIdx == -1 {
		return nil, errors.New("missing zero byte after filename")
	}

	filename := string(filenameBytes)
	packet.Filename = filename

	var modeBytes []byte
	endModeIdx := -1 // the zero delimiter idx after a string.
	for idx, by := range data[endFilenameIdx:] {
		if by == 0x0 {
			endModeIdx = idx
			break
		}
		modeBytes = append(modeBytes, by)
	}

	if endModeIdx == -1 {
		return nil, errors.New("missing zero byte after mode")
	}

	mode := string(modeBytes)
	packet.Mode = mode

	return packet, nil
}

func parseWriteRequest(data []byte) (Packet, error) {

}

func parseAckRequest(data []byte) (Packet, error) {

}

func parseDataRequest(data []byte) (Packet, error) {

}

func parseErrorRequest(data []byte) (Packet, error) {

}
