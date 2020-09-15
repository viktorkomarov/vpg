package auth

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
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

func PasswordMsg(payload string) []byte {
	dst := make([]byte, 5)
	dst[0] = 'p'
	binary.BigEndian.PutUint32(dst[1:5], uint32(len(payload)+5))
	dst = append(dst, password...)
	dst = append(dst, '\000')

	return dst
}

func AuthClient(authentication AuthenticationResponse, user, password, string) Authorizationer {
	switch authentication.Type {
	case AuthenticationMD5Password:
		return authMD5{
			salt: string(authentication.Payload),
			user: user,
			password, password
		}
	case 	
	}
}
