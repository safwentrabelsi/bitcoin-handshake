package main

import (
	"fmt"
	"net"
	"time"

	"github.com/safwentrabelsi/bitcoin-handshake/config"
	"github.com/safwentrabelsi/bitcoin-handshake/network"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	log.SetLevel(logrus.DebugLevel)

	address := net.JoinHostPort(config.BTCNodeHost, fmt.Sprint(config.BTCNodePort))
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	network.ConnectAndHandshake(conn)
}
