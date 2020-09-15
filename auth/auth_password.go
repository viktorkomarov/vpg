package auth

import (
	"fmt"
	"net"
)

type simpleAuth struct {
	password string
}

func (a simpleAuth) Authorize(conn net.Conn) error {
	_, err := conn.Write(PasswordMsg(a.password))
	if err != nil {
		return fmt.Errorf("can't authorize conn %w", err)
	}

	var authentication AuthenticationResponse
	resp := make([]byte, 1024)
	_, err = conn.Read(resp)
	if err != nil {
		return fmt.Errorf("can't authorize conn %w", err)
	}

	if err = authentication.Encode(resp); err != nil {
		return fmt.Errorf("can't encode response msg %w", err)
	}

	if !authentication.Success() {
		return fmt.Errorf("can't authorize %s", authentication.Payload)
	}

	return nil
}
