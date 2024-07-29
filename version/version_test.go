package version

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	"github.com/safwentrabelsi/bitcoin-handshake/config"
	"github.com/safwentrabelsi/bitcoin-handshake/netaddr"
	"github.com/safwentrabelsi/bitcoin-handshake/utils"
	"github.com/stretchr/testify/assert"
)

func TestMakeVersionPayload(t *testing.T) {
	// Setup test config

	payload, err := MakeVersionPayload()
	assert.NoError(t, err, "MakeVersionPayload should not return an error")

	var versionMsg VersionMessage
	reader := bytes.NewReader(payload)

	assert.NoError(t, binary.Read(reader, binary.LittleEndian, &versionMsg.Version))
	assert.Equal(t, int32(config.ProtocolVersion), versionMsg.Version, "Version should match config")

	assert.NoError(t, binary.Read(reader, binary.LittleEndian, &versionMsg.Services))
	assert.Equal(t, uint64(config.Services), versionMsg.Services, "Services should match config")

	assert.NoError(t, binary.Read(reader, binary.LittleEndian, &versionMsg.Timestamp))
	expectedTimestamp := time.Now().Unix()
	assert.InDelta(t, expectedTimestamp, versionMsg.Timestamp, 2, "Timestamp should be within 2 seconds of now")

	assert.NoError(t, netaddr.ParseNetAddr(reader, &versionMsg.AddrRecv))
	assert.NoError(t, netaddr.ParseNetAddr(reader, &versionMsg.AddrFrom))

	assert.NoError(t, binary.Read(reader, binary.LittleEndian, &versionMsg.Nonce))
	assert.Equal(t, uint64(config.NodeID), versionMsg.Nonce, "Nonce should match config")

	var userAgentLen uint8
	assert.NoError(t, binary.Read(reader, binary.LittleEndian, &userAgentLen))
	userAgent := make([]byte, userAgentLen)
	assert.NoError(t, binary.Read(reader, binary.LittleEndian, &userAgent))
	assert.Equal(t, config.UserAgent, string(userAgent), "UserAgent should match config")

	assert.NoError(t, binary.Read(reader, binary.LittleEndian, &versionMsg.StartHeight))
	assert.Equal(t, int32(config.StartHeight), versionMsg.StartHeight, "StartHeight should match config")

	var relay uint8
	assert.NoError(t, binary.Read(reader, binary.LittleEndian, &relay))
	assert.Equal(t, uint8(0), relay, "Relay should be 0")
}

func TestWriteMessageHeader(t *testing.T) {
	// Setup test data
	command := "version"
	payload := []byte{0x01, 0x02, 0x03, 0x04}

	var buf bytes.Buffer
	err := WriteMessageHeader(&buf, command, payload)
	assert.NoError(t, err, "WriteMessageHeader should not return an error")

	var magic [4]byte
	assert.NoError(t, binary.Read(&buf, binary.LittleEndian, &magic))
	assert.Equal(t, config.MainnetMagicBytes, magic, "Magic bytes should match config")

	var commandBytes [12]byte
	assert.NoError(t, binary.Read(&buf, binary.LittleEndian, &commandBytes))
	assert.Equal(t, command, string(bytes.TrimRight(commandBytes[:], "\x00")), "Command should match")

	var length uint32
	assert.NoError(t, binary.Read(&buf, binary.LittleEndian, &length))
	assert.Equal(t, uint32(len(payload)), length, "Payload length should match")

	var checksum [4]byte
	assert.NoError(t, binary.Read(&buf, binary.LittleEndian, &checksum))
	expectedChecksum := utils.CalculateChecksum(payload)
	assert.Equal(t, expectedChecksum[:4], checksum[:], "Checksum should match")
}
