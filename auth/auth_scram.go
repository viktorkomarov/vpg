package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
)

const saslAuthenticationProtocol = "SCRAM-SHA-256"

type scramAuth struct {
	user       string
	password   string
	mechanisms []string
}

func newScramAuth(user, password, serverMechanism string) scramAuth {
	return scramAuth{
		user:       user,
		password:   password,
		mechanisms: strings.Split(serverMechanism, ","),
	}
}

func (a scramAuth) clientFirstMessage() ([]byte, error) {
	buf := make([]byte, 18)

	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}

	result := make([]byte, base64.RawStdEncoding.EncodedLen(len(buf)))
	base64.RawStdEncoding.Encode(result, buf)
	return result, nil
}

func (a scramAuth) containSupportMechanism() bool {
	for _, mechanism := range a.mechanisms {
		if mechanism == saslAuthenticationProtocol {
			return true
		}
	}

	return false
}

func (a scramAuth) Authorize(conn net.Conn) error {
	if !a.containSupportMechanism() { // mv to constructor
		return fmt.Errorf("server doesn't support %s", saslAuthenticationProtocol)
	}

	return nil
}
