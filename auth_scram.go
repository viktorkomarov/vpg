package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"golang.org/x/text/secure/precis"

	"golang.org/x/crypto/pbkdf2"
)

const saslAuthenticationProtocol = "SCRAM-SHA-256"

type scramAuth struct {
	password            []byte
	clientFirst         []byte
	serverFirstResponse []byte
	clientWithoutProof  []byte
	clientFinal         []byte
}

type SASLContinue struct {
	i int
	r []byte
	s []byte
}

func (s *SASLContinue) IsMessage() {}

func newScramAuth(password, mechanisms string) (*scramAuth, error) {
	a := &scramAuth{}

	if !strings.Contains(mechanisms, saslAuthenticationProtocol) {
		return nil, fmt.Errorf("server doesn't support %s", saslAuthenticationProtocol)
	}

	var err error
	a.password, err = precis.OpaqueString.Bytes([]byte(password))
	if err != nil {
		return nil, errors.New("can't prepare password")
	}

	return a, nil
}

//add check that client-first-message contained
func NewSASSLContinue(data []byte) (*SASLContinue, error) {
	serversPayload := bytes.Split(data, []byte(","))
	result := make(map[rune][]byte)

	for _, payload := range serversPayload {
		if len(payload) < 2 {
			return nil, fmt.Errorf("can't define payload type %s", payload)
		}

		result[rune(payload[0])] = payload[2:]
	}

	var err error
	s := &SASLContinue{}
	if s.i, err = strconv.Atoi(string(result['i'])); err != nil {
		return nil, err
	}

	if s.r = result['r']; s.r != nil {
		return nil, errors.New("postgres doesn't send r")
	}

	if s.r = result['s']; s.s != nil {
		return nil, errors.New("postgres doesn't send r")
	}

	return s, nil
}

//refactoring
func (a *scramAuth) clientFirstMessage() ([]byte, error) {
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}

	encoded := make([]byte, base64.RawStdEncoding.EncodedLen(len(buf)))
	base64.RawStdEncoding.Encode(encoded, buf)
	a.clientFirst = []byte(fmt.Sprintf("n=,r=%s", encoded))
	payload := fmt.Sprintf("n,,n=,r=%s", encoded)

	// new func
	result := make([]byte, 5)
	result[0] = 'p'
	result = append(result, []byte(saslAuthenticationProtocol)...)
	s := len(result)
	result = append(result, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(result[s:s+4], uint32(len(payload)))
	result = append(result, payload...)
	binary.BigEndian.PutUint32(result[1:5], uint32(len(result)))

	return result, nil
}

func (a *scramAuth) clientFinalMessage(payload map[rune][]byte) []byte {
	a.clientWithoutProof = []byte(fmt.Sprintf("c=biws,r=%s", payload['r']))
	saltedPassword := pbkdf2.Key(a.password, payload['s'], iter, 32, sha256.New)
	clientKey := a.HMAC(saltedPassword, []byte("Client Key"))
	stroredKey := sha256.Sum256(clientKey)

	authMessage := bytes.Join([][]byte{a.clientFirst, a.serverFirstResponse, a.clientWithoutProof}, []byte(","))
	clientSignature := a.HMAC(stroredKey[:], authMessage)

	clientProof := make([]byte, len(clientSignature))
	for i := range clientProof {
		clientProof[i] = clientKey[i] ^ clientSignature[i]
	}

	buf := make([]byte, base64.StdEncoding.EncodedLen(len(clientProof)))
	base64.StdEncoding.Encode(buf, clientProof)

	msg := fmt.Sprintf("%s,p=%s", a.clientWithoutProof, buf)
	result := []byte{'p', 0, 0, 0, 0}
	binary.BigEndian.PutUint32(result[1:5], uint32(len(msg)+4))

	result = append(result, msg...)

	return result
}

func (a *scramAuth) HMAC(key, msg []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	return mac.Sum(nil)
}

func (a *scramAuth) Authorize(conn net.Conn) error {

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

	serverResponse, err := a.validateServerResponse(authentication.Payload)
	if err != nil {
		return err
	}

	if _, err = conn.Write(a.clientFinalMessage(serverResponse)); err != nil {
		return err
	}

	if _, err = conn.Read(buf); err != nil {
		return err
	}

	log.Fatalf("%s\n", buf)
	if err = authentication.Decode(buf); err != nil {
		return fmt.Errorf("can't decode authenctication %w", err)
	}

	return nil
}