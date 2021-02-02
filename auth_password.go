package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
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
}

func (a authPassword) encode() []byte {
	var b bytes.Buffer

	b.WriteByte('p')
	binary.Write(&b, binary.BigEndian, uint32(len(a.password)+5))
	b.WriteString(a.password)
	b.WriteByte('\000')

	return b.Bytes()
}

type passwordClient struct {
	pswd authPassword
}

func (c passwordClient) authorize(w *Writer, r *Reader) error {
	err := w.sendMsg(c.pswd)
	if err != nil {
		return err
	}

	msg, err := r.receive()
	if err != nil {
		return err
	}

	authOk, ok := msg.(auth)
	if !ok {
		return errors.New("unexpected msg")
	}

	if authOk.Type != authenticationOK {
		return errors.New("unathorized")
	}

	return nil
}
