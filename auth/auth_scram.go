package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
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
	length := 9 + len(encodedInit) + len(saslAuthenticationProtocol)

	binary.BigEndian.PutUint32(result[1:5], uint32(length))
	result = append(result, []byte(saslAuthenticationProtocol)...)
	result = append(result, 0)
	s := len(result)
	result = append(result, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(result[s:s+4], uint32(len(encodedInit)))
	result = append(result, encodedInit...)

	return result, nil
}

func (a *scramAuth) clientFinalMessage(payload map[rune][]byte) ([]byte, error) {
	result := []byte{'p', 0, 0, 0, 0}

	result = append(result, ("c=biws,r=")...)
	result = append(result, payload['r']...)

	iter, err := strconv.Atoi(string(payload['i']))
	if err != nil {
		return nil, err
	}

	encryption := pbkdf2.Key([]byte(a.password), payload['s'], iter, 32, sha256.New)

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

	buf := make([]byte, 4096)
	_, err = conn.Read(buf)
	if err != nil {
		return fmt.Errorf("can't read from connection %w", err)
	}

	var authentication AuthenticationResponse
	if err = authentication.Decode(buf); err != nil {
		return fmt.Errorf("can't decode authenctication %w", err)
	}

	serversResponse, err := validateServerResponse(authentication.Payload)
	if err != nil {
		return err
	}

	return nil
}

func (a *scramAuth) containSupportMechanism() bool {
	return strings.Contains(a.mechanisms, saslAuthenticationProtocol)
}

//add check that client-first-message contained
func (a *scramAuth) validateServerResponse(response []byte) (map[rune][]byte, error) {
	serversPayload := bytes.Split(response, []byte(","))
	result := make(map[rune][]byte)

	for _, payload := range serversPayload {
		if len(payload) < 2 {
			return nil, fmt.Errorf("can't define payload type %s", payload)
		}

		result[rune(payload[0])] = payload[2:]
	}

	for _, t := range []rune{'r', 's', 'i'} {
		if result[t] == nil {
			return nil, fmt.Errorf("server response doesn't contain %c", t)
		}
	}

	return result, nil
}
