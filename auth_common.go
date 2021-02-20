package vpg

type authenticationResponseType uint32

const (
	authenticationOK                authenticationResponseType = 0 // done
	authenticationKerberosV5        authenticationResponseType = 2 // #not_supported : https://postgrespro.ru/docs/postgrespro/10/protocol-flow
	authenticationCleartextPassword authenticationResponseType = 3 // done
	authenticationMD5Password       authenticationResponseType = 5 // done
	authenticationSCMCredential     authenticationResponseType = 6 // unix local only
	authenticationGSS               authenticationResponseType = 7
	authenticationGSSContinue       authenticationResponseType = 8
	authenticationSSPI              authenticationResponseType = 9
	authenticationSASL              authenticationResponseType = 10 // done
	authenticationSASLContinue      authenticationResponseType = 11 // done
	authenticationSASLFinal         authenticationResponseType = 12 // done
)

type auth struct {
	Type    authenticationResponseType
	Payload []byte
}

func (a auth) isMessage() {}

func authClient(msg auth, user, password string) authorizationer {
	switch msg.Type {
	case authenticationMD5Password:
		return passwordClient{
			pswd: authPassword{md5Hash(password, user, string(msg.Payload))},
		}
	case authenticationCleartextPassword:
		return passwordClient{
			pswd: authPassword{password},
		}
		/*case authenticationSASL:
		return scramAuth{
			password:         []byte(password),
			serverMechanisms: string(msg.Payload),
		}*/
	}

	return nil
}
