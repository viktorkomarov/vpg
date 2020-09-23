package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
)

func md5Hash(password, user, salt string) string {
	m := md5.New()
	m.Write([]byte(password + user))
	user = hex.EncodeToString(m.Sum(nil))
	m.Reset()
	m.Write([]byte(user + salt))

	return "md5" + hex.EncodeToString(m.Sum(nil))
}

type authPassword struct {
	password []byte
}

func (a *authPassword) IsMessage() {}

func (a *authPassword) msg() []byte {
	dst := make([]byte, 5)
	dst[0] = 'p'
	binary.BigEndian.PutUint32(dst[1:5], uint32(len(a.Password)+5))
	dst = append(dst, a.Password...)
	dst = append(dst, '\000')

	return dst
}

func (a *authPassword) Authorize(conn net.Conn) error {
	_, err := conn.Write(a.msg())
	if err != nil {
		return fmt.Errorf("can't authorize conn %w", err)
	}

	var authentication ClassificatorAuth
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
