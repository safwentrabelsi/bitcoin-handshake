package network

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/safwentrabelsi/bitcoin-handshake/config"
	"github.com/safwentrabelsi/bitcoin-handshake/netaddr"
	"github.com/safwentrabelsi/bitcoin-handshake/version"
	log "github.com/sirupsen/logrus"
)

const headerLength = 24

type Message struct {
	Command string
	Payload []byte
}

func ConnectAndHandshake(address string) {
	sendChannel := make(chan Message)
	receiveChannel := make(chan []byte)
	doneChannel := make(chan struct{})

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	verackSent := false
	verackReceived := false

	go readMessages(conn, receiveChannel, doneChannel)
	go sendMessages(conn, sendChannel, &verackSent, doneChannel)

	// Send initial version message
	sendChannel <- Message{Command: "version", Payload: createVersionPayload()}

	for {
		select {
		case data := <-receiveChannel:
			err := parseMessage(data, sendChannel, &verackReceived, doneChannel)
			if err != nil {
				log.Errorf("Failed to parse message: %v", err)
				return
			}
			if verackSent && verackReceived {
				log.Info("Handshake completed successfully. Closing connection.")
				close(doneChannel)
				return
			}
		case <-doneChannel:
			return
		}
	}
}

func readMessages(conn net.Conn, receiveChannel chan<- []byte, doneChannel <-chan struct{}) {
	for {
		select {
		case <-doneChannel:
			close(receiveChannel)
			return
		default:
			// Read the header
			header := make([]byte, headerLength)
			_, err := conn.Read(header)
			if err != nil {
				log.Errorf("Failed to read header from connection: %v", err)
				close(receiveChannel)
				return
			}

			// Parse the length from the header
			var length uint32
			headerReader := bytes.NewReader(header[16:20])
			if err := binary.Read(headerReader, binary.LittleEndian, &length); err != nil {
				log.Errorf("Failed to parse length from header: %v", err)
				close(receiveChannel)
				return
			}

			// Read the payload based on the length
			payload := make([]byte, length)
			_, err = conn.Read(payload)
			if err != nil {
				log.Errorf("Failed to read payload from connection: %v", err)
				close(receiveChannel)
				return
			}

			// Send the complete message (header + payload) to the receiveChannel
			completeMessage := append(header, payload...)
			receiveChannel <- completeMessage
		}
	}
}

func sendMessages(conn net.Conn, sendChannel <-chan Message, verackSent *bool, doneChannel <-chan struct{}) {
	for {
		select {
		case msg := <-sendChannel:
			var buf bytes.Buffer
			if err := writeMessageHeader(&buf, msg.Command, msg.Payload); err != nil {
				log.Errorf("Failed to write message header: %v", err)
				continue
			}
			if _, err := buf.Write(msg.Payload); err != nil {
				log.Errorf("Failed to write message payload: %v", err)
				continue
			}
			if _, err := conn.Write(buf.Bytes()); err != nil {
				log.Errorf("Failed to send message: %v", err)
				continue
			}
			log.Infof("Sent %s message", msg.Command)
			if msg.Command == "verack" {
				*verackSent = true
			}
		case <-doneChannel:
			return
		}
	}
}

func parseMessage(data []byte, sendChannel chan<- Message, verackReceived *bool, doneChannel chan<- struct{}) error {
	if len(data) < headerLength {
		return fmt.Errorf("data too short: expected at least %d bytes, got %d", headerLength, len(data))
	}

	header := data[:headerLength]
	payload := data[headerLength:]
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

	expectedMagic := config.MainnetMagicBytes
	if magic != expectedMagic {
		return fmt.Errorf("invalid magic bytes: expected %x, got %x", expectedMagic, magic)
	}

	calculatedChecksum := calculateChecksum(payload[:length])
	if checksum != calculatedChecksum {
		return fmt.Errorf("invalid checksum: expected %x, got %x", checksum, calculatedChecksum)
	}

	if uint32(len(payload)) != length {
		return fmt.Errorf("invalid payload length: expected %d, got %d", length, len(payload))
	}

	trimmedCommand := strings.TrimRight(string(command[:]), "\x00")
	log.Infof("Received %s message. Checksum is valid.", trimmedCommand)
	switch trimmedCommand {
	case "version":
		if err := parseVersionPayload(payload); err != nil {
			return err
		}
		sendChannel <- Message{Command: "verack", Payload: []byte{}}
	case "verack":
		*verackReceived = true
	case "wtxidrelay", "sendaddrv2":
	default:
		if !*verackReceived {
			log.Errorf("Received unknown command: %s. Closing connection.", trimmedCommand)
			close(doneChannel)
			return fmt.Errorf("unknown command received: %s", trimmedCommand)
		}

	}
	return nil
}

func parseVersionPayload(payload []byte) error {
	payloadReader := bytes.NewReader(payload)
	var versionMsg version.VersionMessage
	if err := binary.Read(payloadReader, binary.LittleEndian, &versionMsg.Version); err != nil {
		return err
	}
	if err := binary.Read(payloadReader, binary.LittleEndian, &versionMsg.Services); err != nil {
		return err
	}
	if err := binary.Read(payloadReader, binary.LittleEndian, &versionMsg.Timestamp); err != nil {
		return err
	}
	var addrRecv netaddr.NetAddr
	if err := netaddr.ParseNetAddr(payloadReader, &addrRecv); err != nil {
		return err
	}
	versionMsg.AddrRecv = addrRecv
	var addrFrom netaddr.NetAddr
	if err := netaddr.ParseNetAddr(payloadReader, &addrFrom); err != nil {
		return err
	}
	versionMsg.AddrFrom = addrFrom

	if err := binary.Read(payloadReader, binary.LittleEndian, &versionMsg.Nonce); err != nil {
		return err
	}

	var userAgentLen uint8
	if err := binary.Read(payloadReader, binary.LittleEndian, &userAgentLen); err != nil {
		return err
	}
	userAgent := make([]byte, userAgentLen)
	if err := binary.Read(payloadReader, binary.LittleEndian, &userAgent); err != nil {
		return err
	}
	versionMsg.UserAgent = string(userAgent)

	if strings.Contains(versionMsg.UserAgent, "satoshi") {
		return fmt.Errorf("invalid user agent: expected %s, got %s", config.UserAgent, versionMsg.UserAgent)
	}

	if err := binary.Read(payloadReader, binary.LittleEndian, &versionMsg.StartHeight); err != nil {
		return err
	}

	var relay uint8
	if err := binary.Read(payloadReader, binary.LittleEndian, &relay); err != nil {
		return err
	}
	versionMsg.Relay = (relay != 0)

	log.Debugf("Version: %d", versionMsg.Version)
	log.Debugf("Services: %d", versionMsg.Services)
	log.Debugf("Timestamp: %d", versionMsg.Timestamp)
	log.Debugf("AddrRecv: %+v", versionMsg.AddrRecv)
	log.Debugf("AddrFrom: %+v", versionMsg.AddrFrom)
	log.Debugf("Nonce: %d", versionMsg.Nonce)
	log.Debugf("UserAgent: %s", versionMsg.UserAgent)
	log.Debugf("StartHeight: %d", versionMsg.StartHeight)
	log.Debugf("Relay: %t", versionMsg.Relay)

	return nil
}

func calculateChecksum(payload []byte) [4]byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	var checksum [4]byte
	copy(checksum[:], secondHash[:4])
	return checksum
}

func createVersionPayload() []byte {
	payload, err := version.MakeVersionPayload()
	if err != nil {
		log.Fatalf("Failed to create version payload: %v", err)
	}
	return payload
}

func writeMessageHeader(buf *bytes.Buffer, command string, payload []byte) error {
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

	checksum := calculateChecksum(payload)
	if _, err := buf.Write(checksum[:]); err != nil {
		return err
	}

	return nil
}
