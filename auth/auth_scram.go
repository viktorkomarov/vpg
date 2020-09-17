package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"
)

const saslAuthenticationProtocol = "SCRAM-SHA-256"

type scramAuth struct {
	user       string
	password   string
	mechanisms string
}

func newScramAuth(user, password, serverMechanism string) *scramAuth {
	return &scramAuth{
		user:       user,
		password:   password,
		mechanisms: serverMechanism,
	}
}

//refactoring
func (a *scramAuth) clientFirstMessage() ([]byte, error) {
	result := []byte{'p', 0, 0, 0, 0}

	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	encoded := make([]byte, base64.RawStdEncoding.EncodedLen(len(buf)))
	base64.RawStdEncoding.Encode(encoded, buf)
	encodedInit := fmt.Sprintf("n,,n=,r=%s", encoded)
	length := 6 + len(encodedInit) + len(saslAuthenticationProtocol)
	log.Printf("%s\n", encodedInit)
	binary.BigEndian.PutUint32(result[1:5], uint32(length))
	result = append(result, saslAuthenticationProtocol...)
	result = append(result, '\000')
	result = append(result, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(result[len(result)-4:], uint32(len(encodedInit)))
	result = append(result, encodedInit...)

	return result, nil
}

func (a *scramAuth) Authorize(conn net.Conn) error {
	if !a.containSupportMechanism() { // mv to constructor
		return fmt.Errorf("server doesn't support %s", saslAuthenticationProtocol)
	}

	initMsg, err := a.clientFirstMessage()
	if err != nil {
		return fmt.Errorf("can't create init msg %w", err)
	}

	if _, err := conn.Write(initMsg); err != nil {
		return fmt.Errorf("can't send init msg %w", err)
	}

	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatalf("%s\n", buf)
	return nil
}

func (a *scramAuth) containSupportMechanism() bool {
	return strings.Contains(a.mechanisms, saslAuthenticationProtocol)
}
