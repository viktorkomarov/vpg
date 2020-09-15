package message

import (
	"encoding/binary"
)

const (
	protocolVersion            = uint32(196608)
	saslAuthenticationProtocol = "SCRAM-SHA-256"
)

func StartUpMsg(cfg map[string]string, dst []byte) []byte {
	dst = dst[:0]

	dst = append(dst, 0, 0, 0, 0, 0, 0, 0, 0)
	for key, val := range cfg {
		dst = append(dst, key...)
		dst = append(dst, '\000')
		dst = append(dst, val...)
		dst = append(dst, '\000')
	}

	binary.BigEndian.PutUint32(dst[4:8], protocolVersion)
	dst = append(dst, '\000')
	binary.BigEndian.PutUint32(dst[0:4], uint32(len(dst)))

	return dst
}
