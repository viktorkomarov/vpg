package main

import (
	"fmt"
	"net"

	"gitlab.com/VictorKomarov/vpg/auth"
	"gitlab.com/VictorKomarov/vpg/message"
)

type Conn struct {
	address string
	cfg     map[string]string
	conn    net.Conn
	buf     []byte
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

	if err = c.init(); err != nil {
		c.conn.Close()
		return nil, err
	}

	return c, nil
}

func (c *Conn) init() error {
	c.buf = message.StartUpMsg(c.cfg, c.buf)

	_, err := c.conn.Write(c.buf)
	if err != nil {
		return fmt.Errorf("can't send startup msg %w", err)
	}

	resp := make([]byte, 2048)
	_, err = c.conn.Read(resp) // hate it
	if err != nil {
		return fmt.Errorf("can't read msg %w", err)
	}

	var authentication auth.AuthenticationResponse
	if err := authentication.Encode(resp); err != nil {
		return fmt.Errorf("can't encode authentication msg %w", err)
	}

	if !authentication.Success() {
		client := auth.AuthClient(authentication, c.cfg["user"], c.cfg["password"])
		if err := client.Authorize(c.conn); err != nil {
			return fmt.Errorf("can't authorize connection %w", err)
		}
	}

	return nil
}
