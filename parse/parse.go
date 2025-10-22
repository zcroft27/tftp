package tftp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

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

	return packet, nil
}

func parseReadRequest(data []byte) (Packet, error) {
	isRRQ := true
	return parseReadWriteRequest(data, isRRQ)
}

func parseWriteRequest(data []byte) (Packet, error) {
	isRRQ := false
	return parseReadWriteRequest(data, isRRQ)
}

func parseReadWriteRequest(data []byte, isRRQ bool) (Packet, error) {
	if len(data) < 4 {
		return nil, errors.New("WRQ/RRQ packet is missing opcode and/or required delimiters")
	}

	restPastOpcode := data[2:]

	var filenameBytes []byte
	endFilenameIdx := -1 // the zero delimiter idx after a string.
	for idx, by := range restPastOpcode {
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

	var modeBytes []byte
	endModeIdx := -1 // the zero delimiter idx after a string.
	for idx, by := range restPastOpcode[endFilenameIdx+1:] {
		if by == 0x0 {
			endModeIdx = idx
			break
		}
		modeBytes = append(modeBytes, by)
	}

	if endModeIdx == -1 {
		return nil, errors.New("missing zero byte after mode")
	}

	mode := strings.ToLower(string(modeBytes))
	if _, exists := VALID_MODES[mode]; !exists {
		return nil, errors.New("invalid mode, must be one of: netascii, octet, or mail")
	}

	if isRRQ {
		return ReadRequest{Filename: filename, Mode: mode}, nil
	} else {
		return WriteRequest{Filename: filename, Mode: mode}, nil
	}
}

func parseAckRequest(data []byte) (Packet, error) {
	if len(data) < 4 {
		return nil, errors.New("ACK packet is missing opcode and/or required block number")
	}
	restPastOpcode := data[2:]

	var packet Ack
	blockNumber := binary.BigEndian.Uint16(restPastOpcode[:2])

	packet.BlockNumber = blockNumber

	return packet, nil
}

func parseDataRequest(data []byte) (Packet, error) {
	return nil, errors.New("unimplemented")
}

func parseErrorRequest(data []byte) (Packet, error) {
	return nil, errors.New("unimplemented")
}
