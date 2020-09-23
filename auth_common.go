package main

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

type ClassificatorAuth struct {
	Type    AuthenticationResponseType
	Payload []byte
}

func (c ClassificatorAuth) IsMessage() {}

func AuthClient(authentication ClassificatorAuth, user, password string, conn *Conn) Authorizationer {
	switch authentication.Type {
	case AuthenticationMD5Password:
		return &authPassword{
			password: md5Hash(password, user, string(authentication.Payload)),
			writer:   conn.writer,
		}
	case AuthenticationCleartextPassword:
		return &authPassword{
			password: password,
			writer:   conn.writer,
		}
	case AuthenticationSASL:
		return nil
	}

	return nil
}
