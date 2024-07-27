package netaddr

import (
	"bytes"
	"encoding/binary"
	"net"
)

type NetAddr struct {
	Services uint64
	IP       [16]byte
	Port     uint16
}

func NewNetAddr(ip string, port uint16, services uint64) NetAddr {
	addr := NetAddr{Services: services, Port: port}
	copy(addr.IP[:], net.ParseIP(ip).To16())
	return addr
}

func WriteNetAddr(buf *bytes.Buffer, addr NetAddr) error {
	err := binary.Write(buf, binary.LittleEndian, addr.Services)
	if err != nil {
		return err
	}
	ipBytes := make([]byte, 16)
	copy(ipBytes, addr.IP[:])
	_, err = buf.Write(ipBytes)
	if err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, addr.Port); err != nil {
		return err
	}
	return nil
}

func ParseNetAddr(reader *bytes.Reader, addr *NetAddr) error {
	err := binary.Read(reader, binary.LittleEndian, &addr.Services)
	if err != nil {
		return err
	}
	err = binary.Read(reader, binary.BigEndian, &addr.IP)
	if err != nil {
		return err
	}
	err = binary.Read(reader, binary.BigEndian, &addr.Port)
	if err != nil {
		return err
	}
	return nil
}
