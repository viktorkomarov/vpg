package vpg

import (
	"fmt"
	"net"
	"time"
)

type Conn struct {
	cfg map[string]string

	conn   net.Conn
	reader Reader
	writer Writer

	parametersStatus map[string]string
	pid              int32
	key              int32
	status           byte // may be more cleve
}

func New(cfg map[string]string) (*Conn, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", cfg["address"], time.Second*5)
	if err != nil {
		return nil, fmt.Errorf("can't open connection %w", err)
	}

	c := &Conn{
		conn:             conn,
		cfg:              cfg,
		writer:           newWriter(conn),
		parametersStatus: make(map[string]string),
	}

	c.reader = newReader(c.conn, c)
	if err = c.init(); err != nil {
		c.conn.Close()
		return nil, err
	}

	return c, nil
}

func (c *Conn) init() error {
	if err := c.writer.sendMsg(startUpMsg{
		Version:  protocolVersion,
		User:     c.cfg["user"],
		Database: c.cfg["database"],
	}); err != nil {
		return err
	}

	err := c.authorize()
	if err != nil {
		return err
	}

	for {
		msg, err := c.receive()
		if err != nil {
			return err
		}

		switch msg := msg.(type) {
		case backendKeyData:
			c.pid = msg.PID
			c.key = msg.Key
		case readyForQuery:
			c.status = byte(msg)
			return nil
		default:
			return fmt.Errorf("unexpected msg %+v", msg)
		}
	}
}

func (c *Conn) Close() error {
	return nil
}

func (c *Conn) receive() (message, error) {
	for {
		msg, err := c.reader.receive()
		if err != nil {
			return nil, err
		}

		if status, ok := msg.(parameterStatus); ok {
			c.parametersStatus[status.Name] = status.Value
			continue
		}

		return msg, nil
	}
}

type authorizationer interface {
	authorize(writer *Writer, reader *Reader) error
}

func (c *Conn) authorize() error {
	msg, err := c.receive()
	if err != nil {
		return err
	}

	authType, ok := msg.(auth)
	if !ok {
		return fmt.Errorf("unexpected msg %+v", msg)
	}

	client := authClient(authType, c.cfg["user"], c.cfg["password"])
	if err = client.authorize(c.writer, c.reader); err != nil {
		return err
	}

	return nil
}

func validateConfig(cfg map[string]string) error {
	requiredFields := []string{"address", "user", "password", "database"}

	for _, field := range requiredFields {
		if _, ok := cfg[field]; !ok {
			return fmt.Errorf("required field %s is empty", field)
		}
	}

	return nil
}
