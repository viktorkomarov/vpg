package message

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	protocolVersion            = uint32(196608)
	saslAuthenticationProtocol = "SCRAM-SHA-256"
)

func MD5(needHash string) string {
	m := md5.New()
	m.Write([]byte(needHash))
	return hex.EncodeToString(m.Sum(nil))
}

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

func (a *AuthenticationResponse) Encode(buf []byte) error {
	if len(buf) == 0 {
		return errors.New("empty buf")
	}

	if buf[0] != 'R' {
		return fmt.Errorf("unsupported authencation type %s", buf)
	}

	a.Type = AuthenticationResponseType(binary.BigEndian.Uint32(buf[5:9]))
	a.Payload = buf[9:]

	return nil
}

func SASLMsg(dst []byte, password, mechanism string) []byte {
	password = "n=viktor,r=" + password
	lengthPassword := len(password)
	lengthMessage := 10 + len(saslAuthenticationProtocol) + lengthPassword
	dst = dst[:0]

	dst = append(dst, 0, 0, 0, 0, 0)
	dst[0] = 'p'

	dst = append(dst, saslAuthenticationProtocol...)
	beginLength := len(dst)
	dst = append(dst, 0, 0, 0, 0, 0)

	dst = append(dst, password...)
	dst = append(dst, '\000')

	binary.BigEndian.PutUint32(dst[1:5], uint32(lengthMessage))
	binary.BigEndian.PutUint32(dst[beginLength:beginLength+4], uint32(lengthPassword))
	return dst
}
