package test

import (
	"encoding/binary"
	"math"
	"math/rand"
	"testing"
	tftp "tftp/internal/protocol/parse"

	"github.com/stretchr/testify/assert"
)

const (
	RRQ_UINT16   = uint16(1)
	WRQ_UINT16   = uint16(2)
	DATA_UINT16  = uint16(3)
	ACK_UINT16   = uint16(4)
	ERROR_UINT16 = uint16(5)
)

func TestReadRequest(t *testing.T) {
	for mode := range tftp.VALID_MODES {

		filename := "foo.txt"
		filenameBytes := []byte(filename)
		modeBytes := []byte(mode)
		opCode := tftp.RRQ
		buffer := make([]byte, 2)
		binary.BigEndian.PutUint16(buffer, uint16(opCode))

		assert.Equal(t, RRQ_UINT16, binary.BigEndian.Uint16(buffer))

		readRequest := []byte{}
		readRequest = append(readRequest, buffer...)
		readRequest = append(readRequest, filenameBytes...)
		readRequest = append(readRequest, 0x00)
		readRequest = append(readRequest, modeBytes...)
		readRequest = append(readRequest, 0x00)

		expectedPacket := tftp.ReadRequest{Filename: filename, Mode: mode}
		packet, err := tftp.Parse(readRequest)
		if err != nil {
			t.Error()
		}

		assert.Equal(t, expectedPacket, packet)
	}
}

func TestWriteRequest(t *testing.T) {
	for mode := range tftp.VALID_MODES {

		filename := "foo.txt"
		filenameBytes := []byte(filename)
		modeBytes := []byte(mode)
		opCode := tftp.WRQ
		buffer := make([]byte, 2)
		binary.BigEndian.PutUint16(buffer, uint16(opCode))

		assert.Equal(t, WRQ_UINT16, binary.BigEndian.Uint16(buffer))

		readRequest := []byte{}
		readRequest = append(readRequest, buffer...)
		readRequest = append(readRequest, filenameBytes...)
		readRequest = append(readRequest, 0x00)
		readRequest = append(readRequest, modeBytes...)
		readRequest = append(readRequest, 0x00)

		expectedPacket := tftp.WriteRequest{Filename: filename, Mode: mode}
		packet, err := tftp.Parse(readRequest)
		if err != nil {
			t.Error()
		}

		assert.Equal(t, expectedPacket, packet)
	}
}

func TestAckMessage(t *testing.T) {
	opCode := tftp.ACK
	opCodeBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(opCodeBuffer, uint16(opCode))
	assert.Equal(t, ACK_UINT16, binary.BigEndian.Uint16(opCodeBuffer))

	blockNumber := uint16(rand.Intn(int(math.Pow(2, 16))))
	blockNumberBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(blockNumberBuffer, uint16(blockNumber))
	assert.Equal(t, blockNumber, binary.BigEndian.Uint16(blockNumberBuffer))

	var ackMessage []byte
	ackMessage = append(ackMessage, opCodeBuffer...)
	ackMessage = append(ackMessage, blockNumberBuffer...)

	expectedPacket := tftp.Ack{BlockNumber: blockNumber}
	packet, err := tftp.Parse(ackMessage)
	if err != nil {
		t.Error()
	}
	assert.Equal(t, expectedPacket, packet)
}

func TestErrorMessage(t *testing.T) {
	opCode := tftp.ERROR
	opCodeBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(opCodeBuffer, uint16(opCode))
	assert.Equal(t, ERROR_UINT16, binary.BigEndian.Uint16(opCodeBuffer))

	testCases := []struct {
		errorCode uint16
		errorMsg  string
	}{
		{0, "Not defined"},
		{1, "File not found"},
		{2, "Access violation"},
		{3, "Disk full or allocation exceeded"},
		{4, "Illegal TFTP operation"},
		{5, "Unknown transfer ID"},
		{6, "File already exists"},
		{7, "No such user"},
	}

	for _, tc := range testCases {
		errorCodeBuffer := make([]byte, 2)
		binary.BigEndian.PutUint16(errorCodeBuffer, tc.errorCode)
		assert.Equal(t, tc.errorCode, binary.BigEndian.Uint16(errorCodeBuffer))

		errorMsgBytes := []byte(tc.errorMsg)

		var errorMessage []byte
		errorMessage = append(errorMessage, opCodeBuffer...)
		errorMessage = append(errorMessage, errorCodeBuffer...)
		errorMessage = append(errorMessage, errorMsgBytes...)
		errorMessage = append(errorMessage, 0x00)

		expectedPacket := tftp.Error{ErrorCode: tc.errorCode, ErrorMsg: tc.errorMsg}
		packet, err := tftp.Parse(errorMessage)
		if err != nil {
			t.Errorf("Failed to parse error packet with code %d: %v", tc.errorCode, err)
		}
		assert.Equal(t, expectedPacket, packet)
	}
}

func TestDataRequest(t *testing.T) {
	opCode := tftp.DATA
	opCodeBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(opCodeBuffer, uint16(opCode))
	assert.Equal(t, DATA_UINT16, binary.BigEndian.Uint16(opCodeBuffer))

	blockNumber := uint16(rand.Intn(int(math.Pow(2, 16))))
	blockNumberBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(blockNumberBuffer, uint16(blockNumber))
	assert.Equal(t, blockNumber, binary.BigEndian.Uint16(blockNumberBuffer))

	data := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}

	var dataMessage []byte
	dataMessage = append(dataMessage, opCodeBuffer...)
	dataMessage = append(dataMessage, blockNumberBuffer...)
	dataMessage = append(dataMessage, data...)

	expectedPacket := tftp.Data{BlockNumber: blockNumber, Data: data}
	packet, err := tftp.Parse(dataMessage)
	if err != nil {
		t.Error()
	}
	assert.Equal(t, expectedPacket, packet)
}
