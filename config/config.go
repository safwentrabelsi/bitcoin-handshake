package config

const (
	ProtocolVersion = 70016
	Services        = 1
	UserAgent       = "/Satoshi:27.1.0/"
	StartHeight     = 0
	NodeID          = 12345
	BTCNodeHost     = "0.0.0.0"
	BTCNodePort     = 8333
	Host            = "0.0.0.0"
	Port            = 8333
)

var MainnetMagicBytes = [4]byte{0xf9, 0xbe, 0xb4, 0xd9}
