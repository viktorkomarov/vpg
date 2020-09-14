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

type AuthenticationResponseType uint32

const (
	AuthenticationOK                AuthenticationResponseType = 0 // done
	AuthenticationKerberosV5        AuthenticationResponseType = 2 // #not_supported : https://postgrespro.ru/docs/postgrespro/10/protocol-flow
	AuthenticationCleartextPassword AuthenticationResponseType = 3 // done
	AuthenticationMD5Password       AuthenticationResponseType = 5 // done
	AuthenticationSCMCredential     AuthenticationResponseType = 6 // unix local only
	AuthenticationGSS               AuthenticationResponseType = 7
	AuthenticationGSSContinue       AuthenticationResponseType = 8
	AuthenticationSSPI              AuthenticationResponseType = 9
	AuthenticationSASL              AuthenticationResponseType = 10
	AuthenticationSASLContinue      AuthenticationResponseType = 11
	AuthenticationSASLFinal         AuthenticationResponseType = 12
)

type AuthenticationResponse struct {
	Type    AuthenticationResponseType
	Payload []byte
}

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

func PasswordMsg(dst []byte, password string) []byte {
	dst = dst[:0]

	dst = append(dst, 0, 0, 0, 0, 0)
	dst[0] = 'p'
	binary.BigEndian.PutUint32(dst[1:5], uint32(len(password)+5))
	dst = append(dst, password...)
	dst = append(dst, '\000')

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

func SaslMsg(dst []byte, password string) []byte {
	dst = dst[:1]

	dst = append(dst, 0, 0, 0, 0, 0)
	dst[0] = 'p'

	dst = append(dst, saslAuthenticationProtocol...)
	dst = append(dst, 0, 0, 0, 0, 0)
	password = "n=,r=" + password

	//do it
}
