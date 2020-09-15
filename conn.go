package main

import (
	"fmt"
	"log"
	"net"

	"gitlab.com/VictorKomarov/vpg/message"
)

type Conn struct {
	address string
	cfg     map[string]string
	conn    net.Conn
}

type Authorizationer interface {
	Authorize(conn net.Conn) error
}

func New(address string) (*Conn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("can't init connection %w", err)
	}

	c := &Conn{
		cfg: map[string]string{
			"user":     "viktor",
			"password": "password",
			"database": "viktor",
		},
		conn: conn,
		buf:  make([]byte, 0, 1024),
	}

	return c, c.init()
}

func (c *Conn) init() error {
	c.buf = message.StartUpMsg(c.cfg, c.buf)

	connected := false
	for !connected {
		_, err := c.conn.Write(c.buf)
		if err != nil {
			return fmt.Errorf("can't send startup msg %w", err)
		}

		resp := make([]byte, 2048)
		_, err = c.conn.Read(resp) // hate it
		if err != nil {
			return fmt.Errorf("can't read msg %w", err)
		}

		var authentication message.AuthenticationResponse
		if err := authentication.Encode(resp); err != nil {
			return fmt.Errorf("can't encode authentication msg %w", err)
		}

		switch authentication.Type {
		case message.AuthenticationOK:
			connected = true
		case message.AuthenticationMD5Password:
			salt := string(authentication.Payload[:4]) // bad
			md5Hashed := "md5" + message.MD5(message.MD5(c.cfg["password"]+c.cfg["user"])+salt)
			c.buf = message.PasswordMsg(c.buf, md5Hashed)
		case message.AuthenticationCleartextPassword:
			password := c.cfg["password"]
			c.buf = message.PasswordMsg(c.buf, password)
		case message.AuthenticationSASL:
			saslMechanism := string(authentication.Payload) // postgresql support only SCRAM-SHA-256
			c.buf = message.SASLMsg(c.buf, c.cfg["password"], saslMechanism)
		default:
			log.Fatalf("%s\n", c.buf)
		}
	}

	return nil
}
