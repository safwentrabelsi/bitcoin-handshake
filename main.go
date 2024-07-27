package main

import (
	"fmt"
	"net"

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

	network.ConnectAndHandshake(net.JoinHostPort(config.BTCNodeHost, fmt.Sprint(config.BTCNodePort)))

}
