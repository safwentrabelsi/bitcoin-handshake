package utils

import "crypto/sha256"

func CalculateChecksum(payload []byte) [4]byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	var checksum [4]byte
	copy(checksum[:], secondHash[:4])
	return checksum
}
