package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	protocolVersion = 70016
	services        = 1
	userAgent       = "/Satoshi:0.21.0/"
	startHeight     = 0
	nodeId          = 12345
)

type VersionMessage struct {
	Version     int32
	Services    uint64
	Timestamp   int64
	AddrRecv    NetAddr
	AddrFrom    NetAddr
	Nonce       uint64
	UserAgent   string
	StartHeight int32
	Relay       bool
}

type NetAddr struct {
	Services uint64
	IP       [16]byte
	Port     uint16
}

func main() {
	// Connect to the Bitcoin node
	log.Println("Connecting to Bitcoin node...")
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8333", 10*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	log.Println("Connected to Bitcoin node")

	// Create and send the version message
	sendVersionMessage(conn)

	// Read the response from the node
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	log.Printf("Received response: %x", buffer[:n])
	if err := parseVersionMessage(buffer[:n]); err != nil {
		log.Fatalf("Failed to parse version message: %v", err)
	}

	// Send the verack message
	log.Println("Sending verack message...")
	sendVerackMessage(conn)

	// Read the verack response
	n, err = conn.Read(buffer)
	if err != nil {
		log.Fatalf("Failed to read verack response: %v", err)
	}

	log.Printf("Received verack response: %x", buffer[:n])
	if err := parseVerackMessage(buffer[:n]); err != nil {
		log.Fatalf("Failed to parse verack message: %v", err)
	}

	log.Println("Handshake completed successfully")
}

func sendVersionMessage(conn net.Conn) {
	var buf bytes.Buffer

	// Construct the version message payload
	payload := makeVersionPayload()
	log.Println("len payload:", len(payload))
	// Write the message header
	writeMessageHeader(&buf, "version", payload)

	// Write the payload
	buf.Write(payload)
	log.Printf("%x\n", buf.Bytes())
	// Send the version message

	_, err := conn.Write(buf.Bytes())
	if err != nil {
		log.Fatalf("Failed to send version message: %v", err)
	}
	log.Println("Version message sent")
}

func makeVersionPayload() []byte {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, int32(protocolVersion))
	binary.Write(&buf, binary.LittleEndian, uint64(services))
	binary.Write(&buf, binary.LittleEndian, uint64(time.Now().Unix()))
	writeNetAddr(&buf, NewNetAddr("0.0.0.0", 8333, services))
	writeNetAddr(&buf, NewNetAddr("0.0.0.0", 8333, services))

	binary.Write(&buf, binary.LittleEndian, uint64(nodeId)) // Random nonce
	buf.WriteByte(byte(len(userAgent)))
	buf.WriteString(userAgent)
	binary.Write(&buf, binary.LittleEndian, int32(startHeight))
	log.Println("payload:", hex.EncodeToString(buf.Bytes()))
	return buf.Bytes()
}

func writeNetAddr(buf *bytes.Buffer, addr NetAddr) {
	binary.Write(buf, binary.LittleEndian, addr.Services)
	ipBytes := make([]byte, 16)
	copy(ipBytes, addr.IP[:])

	buf.Write(ipBytes) // Unused
	binary.Write(buf, binary.BigEndian, addr.Port)
}

func writeMessageHeader(buf *bytes.Buffer, command string, payload []byte) {
	buf.Write([]byte{0xf9, 0xbe, 0xb4, 0xd9}) // Magic bytes
	commandBytes := make([]byte, 12)
	copy(commandBytes, command)
	buf.Write(commandBytes)
	binary.Write(buf, binary.LittleEndian, uint32(len(payload)))

	// Temporarily store current buffer content
	currentContent := buf.Bytes()

	// Create a buffer for checksum calculation
	tempBuf := new(bytes.Buffer)
	tempBuf.Write(currentContent[:16])        // magic + command + length (16 bytes)
	tempBuf.Write(make([]byte, len(payload))) // Add payload length of zeros for checksum calculation

	// Calculate the checksum
	checksum := calculateChecksum(payload)
	buf.Write(checksum)
	log.Println(hex.EncodeToString(buf.Bytes()))
}

func calculateChecksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	log.Println("checksum:", hex.EncodeToString(secondHash[:]))
	return secondHash[:4]
}

func parseVersionMessage(data []byte) error {
	header := data[:24]
	payload := data[24:]

	// Parse the header
	var magic [4]byte
	var command [12]byte
	var length uint32
	var checksum [4]byte

	headerReader := bytes.NewReader(header)
	if err := binary.Read(headerReader, binary.LittleEndian, &magic); err != nil {
		return err
	}
	if err := binary.Read(headerReader, binary.LittleEndian, &command); err != nil {
		return err
	}
	if err := binary.Read(headerReader, binary.LittleEndian, &length); err != nil {
		return err
	}
	if err := binary.Read(headerReader, binary.LittleEndian, &checksum); err != nil {
		return err
	}

	log.Printf("Magic: %x", magic)
	log.Printf("Command: %s", command)
	log.Printf("Length: %d", length)
	log.Printf("Checksum: %x", checksum)

	if string(bytes.Trim(command[:], "\x00")) != "version" {
		return fmt.Errorf("Expected version message, got %s", command)
	}

	// Parse the payload
	payloadReader := bytes.NewReader(payload)
	var versionMsg VersionMessage
	if err := binary.Read(payloadReader, binary.LittleEndian, &versionMsg.Version); err != nil {
		return err
	}
	if err := binary.Read(payloadReader, binary.LittleEndian, &versionMsg.Services); err != nil {
		return err
	}
	if err := binary.Read(payloadReader, binary.LittleEndian, &versionMsg.Timestamp); err != nil {
		return err
	}
	if err := binary.Read(payloadReader, binary.LittleEndian, &versionMsg.AddrRecv); err != nil {
		return err
	}
	if err := binary.Read(payloadReader, binary.LittleEndian, &versionMsg.AddrFrom); err != nil {
		return err
	}
	if err := binary.Read(payloadReader, binary.LittleEndian, &versionMsg.Nonce); err != nil {
		return err
	}

	// Read the user agent
	var userAgentLen uint8
	if err := binary.Read(payloadReader, binary.LittleEndian, &userAgentLen); err != nil {
		return err
	}
	userAgent := make([]byte, userAgentLen)
	if err := binary.Read(payloadReader, binary.LittleEndian, &userAgent); err != nil {
		return err
	}
	versionMsg.UserAgent = string(userAgent)

	if err := binary.Read(payloadReader, binary.LittleEndian, &versionMsg.StartHeight); err != nil {
		return err
	}

	var relay uint8
	if err := binary.Read(payloadReader, binary.LittleEndian, &relay); err != nil {
		return err
	}
	versionMsg.Relay = (relay != 0)

	log.Printf("Version: %d", versionMsg.Version)
	log.Printf("Services: %d", versionMsg.Services)
	log.Printf("Timestamp: %d", versionMsg.Timestamp)
	log.Printf("AddrRecv: %+v", versionMsg.AddrRecv)
	log.Printf("AddrFrom: %+v", versionMsg.AddrFrom)
	log.Printf("Nonce: %d", versionMsg.Nonce)
	log.Printf("UserAgent: %s", versionMsg.UserAgent)
	log.Printf("StartHeight: %d", versionMsg.StartHeight)
	log.Printf("Relay: %t", versionMsg.Relay)

	return nil
}

func sendVerackMessage(conn net.Conn) {
	var buf bytes.Buffer
	writeMessageHeader(&buf, "verack", []byte{})
	// data, err := hex.DecodeString("F9BEB4D976657261636B000000000000000000005DF6E0E2")
	// if err != nil {
	// 	panic(err)
	// }
	_, err := conn.Write(buf.Bytes())
	if err != nil {
		log.Fatalf("Failed to send verack message: %v", err)
	}
	log.Println("Verack message sent")
}

func parseVerackMessage(data []byte) error {
	header := data[:24]
	headerReader := bytes.NewReader(header)

	// Read the header
	var magic [4]byte
	var command [12]byte
	var length uint32
	var checksum [4]byte

	if err := binary.Read(headerReader, binary.LittleEndian, &magic); err != nil {
		return err
	}
	if err := binary.Read(headerReader, binary.LittleEndian, &command); err != nil {
		return err
	}
	if err := binary.Read(headerReader, binary.LittleEndian, &length); err != nil {
		return err
	}
	if err := binary.Read(headerReader, binary.LittleEndian, &checksum); err != nil {
		return err
	}

	log.Printf("Magic: %x", magic)
	log.Printf("Command: %s", command)
	log.Printf("Length: %d", length)
	log.Printf("Checksum: %x", checksum)

	// whatever command no check here
	if string(bytes.Trim(command[:], "\x00")) != "verack" {
		return fmt.Errorf("Expected verack message, got %s", command)
	}

	log.Println("Verack message received")
	return nil
}

func NewNetAddr(ip string, port uint16, services uint64) NetAddr {
	addr := NetAddr{Services: services, Port: port}
	copy(addr.IP[:], net.ParseIP(ip).To16())
	return addr
}
