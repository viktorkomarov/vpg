package auth

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
)

type authMD5 struct {
	salt     string
	user     string
	password string
}

func (a authMD5) message() string {
	m := md5.New()
	m.Write([]byte(a.password + a.user))
	user := hex.EncodeToString(m.Sum(nil))
	m.Reset()
	m.Write([]byte(user + salt))

	return "md5" + hex.EncodeToString(m.Sum(nil))
}

func (a authMD5) Authorize(conn net.Conn) error {
	_, err := conn.Write(PasswordMsg(a.message()))
	if err != nil {
		return fmt.Errorf("can't authorize conn %w", err)
	}

	var authentication AuthenticationResponse
	resp := make([]byte, 1024)
	_, err := conn.Read(resp)
	if err != nil {
		return fmt.Errorf("can't authorize conn %w", err)
	}

	if err := authentication.Encode(resp); err != nil {
		return fmt.Errorf("can't encode response msg %w", err)
	}

	if !authentication.Success() {
		return fmt.Errorf("can't authorize by md5 %s", authentication.Payload)
	}

	return nil
}
