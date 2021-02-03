package main

import (
	"crypto/md5"
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
	return append([]byte(a.password), '\000')
}

type passwordClient struct {
	pswd authPassword
}

func (c passwordClient) authorize(w *Writer, r *Reader) error {
	err := w.mType('p').sendMsg(c.pswd)
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
