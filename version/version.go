package version

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/safwentrabelsi/bitcoin-handshake/config"
	"github.com/safwentrabelsi/bitcoin-handshake/netaddr"
	"github.com/safwentrabelsi/bitcoin-handshake/utils"
)

type VersionMessage struct {
	Version     int32
	Services    uint64
	Timestamp   int64
	AddrRecv    netaddr.NetAddr
	AddrFrom    netaddr.NetAddr
	Nonce       uint64
	UserAgent   string
	StartHeight int32
	Relay       bool
}

func MakeVersionPayload() ([]byte, error) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.LittleEndian, int32(config.ProtocolVersion)); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, uint64(config.Services)); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, uint64(time.Now().Unix())); err != nil {
		return nil, err
	}
	if err := netaddr.WriteNetAddr(&buf, netaddr.NewNetAddr(config.BTCNodeHost, config.BTCNodePort, config.Services)); err != nil {
		return nil, err
	}
	if err := netaddr.WriteNetAddr(&buf, netaddr.NewNetAddr(config.Host, config.Port, config.Services)); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, uint64(config.NodeID)); err != nil {
		return nil, err
	}
	if err := buf.WriteByte(byte(len(config.UserAgent))); err != nil {
		return nil, err
	}
	if _, err := buf.WriteString(config.UserAgent); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, int32(config.StartHeight)); err != nil {
		return nil, err
	}
	if err := buf.WriteByte(0); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func WriteMessageHeader(buf *bytes.Buffer, command string, payload []byte) error {
	if _, err := buf.Write(config.MainnetMagicBytes[:]); err != nil {
		return err
	}

	commandBytes := make([]byte, 12)
	copy(commandBytes, command)
	if _, err := buf.Write(commandBytes); err != nil {
		return err
	}

	if err := binary.Write(buf, binary.LittleEndian, uint32(len(payload))); err != nil {
		return err
	}

	checksum := utils.CalculateChecksum(payload)
	if _, err := buf.Write(checksum[:]); err != nil {
		return err
	}

	return nil
}
