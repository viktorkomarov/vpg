package auth

import (
	"encoding/binary"
	"fmt"
	"net"
)

type simpleAuth struct {
	password string
}

func (a simpleAuth) msg() []byte {
	dst := make([]byte, 5)
	dst[0] = 'p'
	binary.BigEndian.PutUint32(dst[1:5], uint32(len(a.password)+5))
	dst = append(dst, a.password...)
	dst = append(dst, '\000')

	return dst
}

func (a simpleAuth) Authorize(conn net.Conn) error {
	_, err := conn.Write(a.msg())
	if err != nil {
		return fmt.Errorf("can't authorize conn %w", err)
	}

	var authentication AuthenticationResponse
	resp := make([]byte, 1024)
	_, err = conn.Read(resp)
	if err != nil {
		return fmt.Errorf("can't authorize conn %w", err)
	}

	if err = authentication.Decode(resp); err != nil {
		return fmt.Errorf("can't decode response msg %w", err)
	}

	if !authentication.Success() {
		return fmt.Errorf("can't authorize %s", authentication.Payload)
	}

	return nil
}
