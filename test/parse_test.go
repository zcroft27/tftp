package test

import (
	"encoding/binary"
	"testing"
	tftp "tftp/parse"

	"github.com/stretchr/testify/assert"
)

const (
	RRQ_UINT16 = uint16(1)
	WRQ_UINT16 = uint16(2)
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
