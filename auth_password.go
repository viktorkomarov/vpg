package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
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
	password string
	writer   *Writer
}

func (a *authPassword) Encode() []byte {
	dst := make([]byte, 5)
	dst[0] = 'p'
	binary.BigEndian.PutUint32(dst[1:5], uint32(len(a.password)+5))
	dst = append(dst, a.password...)
	dst = append(dst, '\000')

	return dst
}

func (a *authPassword) Authorize() error {
	return a.writer.Send(a)
}
