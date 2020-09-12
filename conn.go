package main

import (
	"fmt"
	"net"

	"gitlab.com/VictorKomarov/vpg/message"
)

type Conn struct {
	address string
	cfg     map[string]string
	conn    net.Conn
	reqBuf  []byte
	resBuf  []byte
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
		conn:   conn,
		reqBuf: make([]byte, 0, 1024),
		resBuf: make([]byte, 1024),
	}

	return c, c.init()
}

func (c *Conn) init() error {
	c.reqBuf = message.StartUpMsg(c.cfg, c.reqBuf)

	connected := false
	for !connected {
		_, err := c.conn.Write(c.reqBuf)
		if err != nil {
			return fmt.Errorf("can't send startup msg %w", err)
		}

		_, err = c.conn.Read(c.resBuf)
		if err != nil {
			return fmt.Errorf("can't read msg %w", err)
		}

		var authentication message.AuthenticationResponse
		if err := authentication.Encode(c.resBuf); err != nil {
			return fmt.Errorf("can't encode authentication msg %w", err)
		}

		switch authentication.Type {
		case message.AuthenticationOK:
			connected = true
		case message.AuthenticationMD5Password:
			salt := string(authentication.Payload[:4]) // bad
			md5Hashed := message.MD5(message.MD5(c.cfg["password"]+c.cfg["user"]) + salt)
			c.reqBuf = message.MD5Msg(c.reqBuf, md5Hashed)
		}
	}

	return nil
}
