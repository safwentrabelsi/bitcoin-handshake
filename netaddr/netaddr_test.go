package netaddr

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
)

func TestNewNetAddr(t *testing.T) {
	ip := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	port := uint16(8333)
	services := uint64(1)

	addr := NewNetAddr(ip, port, services)

	expectedIP := net.ParseIP(ip).To16()
	if !bytes.Equal(addr.IP[:], expectedIP) {
		t.Errorf("Expected IP %v, got %v", expectedIP, addr.IP)
	}

	if addr.Port != port {
		t.Errorf("Expected port %d, got %d", port, addr.Port)
	}

	if addr.Services != services {
		t.Errorf("Expected services %d, got %d", services, addr.Services)
	}
}

func TestWriteNetAddr(t *testing.T) {
	ip := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	port := uint16(8333)
	services := uint64(1)

	addr := NewNetAddr(ip, port, services)

	var buf bytes.Buffer
	err := WriteNetAddr(&buf, addr)
	if err != nil {
		t.Fatalf("Failed to write NetAddr: %v", err)
	}

	var writtenServices uint64
	err = binary.Read(&buf, binary.LittleEndian, &writtenServices)
	if err != nil {
		t.Fatalf("Failed to read services: %v", err)
	}
	if writtenServices != services {
		t.Errorf("Expected services %d, got %d", services, writtenServices)
	}

	writtenIP := make([]byte, 16)
	_, err = buf.Read(writtenIP)
	if err != nil {
		t.Fatalf("Failed to read IP: %v", err)
	}
	expectedIP := net.ParseIP(ip).To16()
	if !bytes.Equal(writtenIP, expectedIP) {
		t.Errorf("Expected IP %v, got %v", expectedIP, writtenIP)
	}

	var writtenPort uint16
	err = binary.Read(&buf, binary.BigEndian, &writtenPort)
	if err != nil {
		t.Fatalf("Failed to read port: %v", err)
	}
	if writtenPort != port {
		t.Errorf("Expected port %d, got %d", port, writtenPort)
	}
}

func TestParseNetAddr(t *testing.T) {
	ip := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	port := uint16(8333)
	services := uint64(1)

	addr := NewNetAddr(ip, port, services)

	var buf bytes.Buffer
	err := WriteNetAddr(&buf, addr)
	if err != nil {
		t.Fatalf("Failed to write NetAddr: %v", err)
	}

	var parsedAddr NetAddr
	err = ParseNetAddr(bytes.NewReader(buf.Bytes()), &parsedAddr)
	if err != nil {
		t.Fatalf("Failed to parse NetAddr: %v", err)
	}

	if parsedAddr.Services != services {
		t.Errorf("Expected services %d, got %d", services, parsedAddr.Services)
	}

	expectedIP := net.ParseIP(ip).To16()
	if !bytes.Equal(parsedAddr.IP[:], expectedIP) {
		t.Errorf("Expected IP %v, got %v", expectedIP, parsedAddr.IP)
	}

	if parsedAddr.Port != port {
		t.Errorf("Expected port %d, got %d", port, parsedAddr.Port)
	}
}
