package auth

import (
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

func (a scramAuth) containSupportMechanism() bool {
	for _, mechanism := range a.mechanisms {
		if mechanism == saslAuthenticationProtocol {
			return true
		}
	}

	return false
}

func (a scramAuth) Authorize(conn net.Conn) error {
	if !a.containSupportMechanism() {
		return fmt.Errorf("server doesn't support %s", saslAuthenticationProtocol)
	}

	return nil
}
