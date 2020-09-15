package auth

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
)

//mv where i use it
type Authorizationer interface {
	Authorize(conn net.Conn) error
}

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

func (a *AuthenticationResponse) Success() bool {
	return a.Type == AuthenticationOK
}

func md5Hash(password, user, salt string) string {
	m := md5.New()
	m.Write([]byte(password + user))
	user = hex.EncodeToString(m.Sum(nil))
	m.Reset()
	m.Write([]byte(user + salt))

	return "md5" + hex.EncodeToString(m.Sum(nil))
}

func AuthClient(authentication AuthenticationResponse, user, password string) Authorizationer {
	switch authentication.Type {
	case AuthenticationMD5Password:
		return simpleAuth{
			password: md5Hash(password, user, string(authentication.Payload)),
		}
	case AuthenticationCleartextPassword:
		return simpleAuth{
			password: password,
		}
	case AuthenticationSASL:
		return newScramAuth(user, password, string(authentication.Payload))
	}

	return nil
}
