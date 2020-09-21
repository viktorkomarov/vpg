package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const saslAuthenticationProtocol = "SCRAM-SHA-256"

type scramAuth struct {
	password   string
	mechanisms string

	clientFirst         []byte
	serverFirstResponse []byte
	clientWithoutProof  []byte
	clientFinal         []byte
}

func newScramAuth(password, serverMechanism string) *scramAuth {
	return &scramAuth{
		password:   password,
		mechanisms: serverMechanism,
	}
}

//add check that client-first-message contained && delete about iter when size msg will standart
func (a *scramAuth) validateServerResponse(response []byte) (map[rune][]byte, error) {
	a.serverFirstResponse = response

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

	result['i'] = result['i'][:4]
	return result, nil
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

	a.clientFirst = []byte(fmt.Sprintf("n=,r=%s", encoded))
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

func (a *scramAuth) clientFinalMessage(payload map[rune][]byte) []byte {
	result := []byte{'p', 0, 0, 0, 0}

	a.clientWithoutProof = []byte(fmt.Sprintf("c=biws,r=%s", string(payload['r'])))
	result = append(result, fmt.Sprintf("%s", string(a.clientWithoutProof))...)

	iter, _ := strconv.Atoi(string(payload['i']))
	saltedPassword := pbkdf2.Key([]byte(a.password), payload['s'], iter, 32, sha256.New)
	clientKey := a.HMAC(saltedPassword, []byte("Client Key"))
	stroredKey := sha256.Sum256(clientKey)

	authMessage := bytes.Join([][]byte{a.clientFinal, a.serverFirstResponse, a.clientWithoutProof}, []byte(","))
	clientSignature := a.HMAC(stroredKey[:], authMessage)

	clientProof := make([]byte, len(clientSignature))
	for i := range clientProof {
		clientProof[i] = clientKey[i] ^ clientSignature[i]
	}

	result = append(result, []byte(fmt.Sprintf(",p=%s", string(clientProof)))...)
	binary.BigEndian.PutUint32(result[1:5], 1024)

	log.Printf("%s _ %d __ %s\n", string(result[0]), binary.BigEndian.Uint32(result[1:5]), result[5:])
	return result
}

func (a *scramAuth) HMAC(key, msg []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	return mac.Sum(nil)
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

func (a *scramAuth) containSupportMechanism() bool {
	return strings.Contains(a.mechanisms, saslAuthenticationProtocol)
}
