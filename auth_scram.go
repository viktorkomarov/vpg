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
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/text/secure/precis"
)

const saslAuthenticationProtocol = "SCRAM-SHA-256"

type scramAuth struct {
	password            []byte
	clientFirst         []byte
	serverFirstResponse *SASLContinue
	clientWithoutProof  []byte
	saltedPassword      []byte
	clientProof         []byte
	authMessage         []byte
	generatedFirstMsg   []byte

	writer     *Writer
	reader     *Reader
	mechanisms string
}

func (a *scramAuth) Authorize() error {
	err := a.prepare()
	if err != nil {
		return err
	}

	clientFirst := ClientFirstMessage{
		Protocol: []byte(saslAuthenticationProtocol),
		Data:     []byte(fmt.Sprintf("n,,%s", a.clientFirst)),
	}

	if err = a.writer.Send(clientFirst); err != nil {
		return err
	}

	if err := a.receiveSASLContinue(); err != nil {
		return err
	}

	a.clientFinalMessage()
	saslResp := SASLResponse{
		Data: []byte(fmt.Sprintf("%s,p=%s", a.clientWithoutProof, a.clientProof)),
	}

	if err = a.writer.Send(saslResp); err != nil {
		return err
	}

	return a.validateSASLFinal()
}

func (a *scramAuth) prepare() error {
	if !strings.Contains(a.mechanisms, saslAuthenticationProtocol) {
		return fmt.Errorf("server doesn't support %s", saslAuthenticationProtocol)
	}

	var err error
	a.password, err = precis.OpaqueString.Bytes(a.password)
	if err != nil {
		return errors.New("can't prepare password")
	}

	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Errorf("can't generate random %w", err)
	}
	encoded := make([]byte, base64.RawStdEncoding.EncodedLen(len(buf)))
	base64.RawStdEncoding.Encode(encoded, buf)
	a.generatedFirstMsg = encoded
	a.clientFirst = []byte(fmt.Sprintf("n=,r=%s", encoded))

	return nil
}

func HMAC(key, msg []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	return mac.Sum(nil)
}

func (a *scramAuth) clientFinalMessage() {
	a.clientWithoutProof = []byte(fmt.Sprintf("c=biws,r=%s", a.serverFirstResponse.r))
	a.saltedPassword = pbkdf2.Key(a.password, a.serverFirstResponse.s, a.serverFirstResponse.i, 32, sha256.New)
	clientKey := HMAC(a.saltedPassword, []byte("Client Key"))
	stroredKey := sha256.Sum256(clientKey)
	a.authMessage = bytes.Join([][]byte{a.clientFirst, a.serverFirstResponse.all, a.clientWithoutProof}, []byte(","))
	clientSignature := HMAC(stroredKey[:], a.authMessage)

	clientProof := make([]byte, len(clientSignature))
	for i := range clientProof {
		clientProof[i] = clientKey[i] ^ clientSignature[i]
	}

	a.clientProof = make([]byte, base64.StdEncoding.EncodedLen(len(clientProof)))
	base64.StdEncoding.Encode(a.clientProof, clientProof)
}

type ClientFirstMessage struct {
	Protocol []byte
	Data     []byte
}

func (c ClientFirstMessage) Encode() []byte {
	result := make([]byte, 5)
	result[0] = 'p'
	result = append(result, c.Protocol...)
	result = append(result, '\000')
	s := len(result)
	result = append(result, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(result[s:s+4], uint32(len(c.Data)))
	result = append(result, c.Data...)
	binary.BigEndian.PutUint32(result[1:5], uint32(len(result)-1))

	return result
}

type SASLContinue struct {
	i   int
	r   []byte
	s   []byte
	all []byte
}

func (s *SASLContinue) IsMessage() {}

func NewSASSLContinue(data []byte) (*SASLContinue, error) {
	s := &SASLContinue{
		all: data,
	}

	serversPayload := bytes.Split(data, []byte(","))
	result := make(map[string][]byte)
	for _, payload := range serversPayload {
		if len(payload) < 2 {
			return nil, fmt.Errorf("can't define payload type %s", payload)
		}

		result[string(payload[0])] = payload[2:]
	}

	var err error
	if s.i, err = strconv.Atoi(string(result["i"])); err != nil {
		return nil, err
	}

	if s.r = result["r"]; s.r == nil {
		return nil, errors.New("postgres doesn't send r")
	}

	if s.s = result["s"]; s.s == nil {
		return nil, errors.New("postgres doesn't send s")
	}

	return s, nil
}

func (s *scramAuth) receiveSASLContinue() error {
	msg, err := s.reader.Receive()
	if err != nil {
		return err
	}

	ser, ok := msg.(*SASLContinue)
	if !ok {
		return errors.New("cast error")
	}

	if !bytes.Contains(ser.r, s.generatedFirstMsg) {
		return fmt.Errorf("uncorrect ser msg %s", ser.r)
	}

	s.serverFirstResponse = ser
	return nil
}

type SASLResponse struct {
	Data []byte
}

func (s SASLResponse) Encode() []byte {
	buf := make([]byte, 5)
	buf[0] = 'p'

	binary.BigEndian.PutUint32(buf[1:5], uint32(len(s.Data)+4))
	buf = append(buf, s.Data...)

	log.Fatalf("%s", buf[5:])
	return buf
}

type SASLFinal struct {
	data []byte
}

func (s SASLFinal) IsMessage() {}

func NewSASLFinal(data []byte) (*SASLFinal, error) {
	if !bytes.Contains(data, []byte("v=")) {
		return nil, fmt.Errorf("uncorrect sasl final %s", data)
	}

	s := &SASLFinal{
		data: data[2:],
	}

	return s, nil
}

func (s *scramAuth) validateSASLFinal() error {
	msg, err := s.reader.Receive()
	if err != nil {
		return err
	}

	final, ok := msg.(*SASLFinal)
	if !ok {
		return fmt.Errorf("expected sasl final %+v", msg)
	}

	serverKey := HMAC(s.saltedPassword, []byte("Server Key"))
	serverSignature := HMAC(serverKey, s.authMessage)
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(serverSignature)))
	base64.StdEncoding.Encode(buf, serverSignature)

	if !hmac.Equal(final.data, buf) {
		return fmt.Errorf("uncorrect server hash %s", final.data)
	}

	return nil
}
