package main

const saslAuthenticationProtocol = "SCRAM-SHA-256"

/*type scramAuth struct {
	reader *Reader
	writer *Writer

	serverMechanisms string

	password                       []byte
	clientFirstMessageBare         []byte
	clientFirstMessage             []byte
	serverFirstMessage             *SASLContinue
	clientFinalMessageWithoutProof []byte
	saltedPassword                 []byte
	authMessage                    []byte
}

func (a *scramAuth) authorize() error {
	err := a.initAuth()
	if err != nil {
		return err
	}

	a.clientFirstMessage = []byte(fmt.Sprintf("n=,r=%s", a.clientFirstMessageBare))
	clientFirst := ClientFirstMessage{
		Protocol: []byte(saslAuthenticationProtocol),
		Data:     []byte(fmt.Sprintf("n,,%s", a.clientFirstMessage)),
	}

	if err = a.writer.send(clientFirst); err != nil {
		return err
	}

	a.serverFirstMessage, err = a.receiveSASLContinue()
	if err != nil {
		return err
	}

	proof := a.clientFinalMessage()
	clientFinal := SASLResponse{
		Data: []byte(proof),
	}

	if err = a.writer.Send(clientFinal); err != nil {
		return err
	}

	msg, err := a.reader.Receive()
	log.Fatalf("%+v \n %+v", msg, err)
	return nil
}

func (a *scramAuth) initAuth() error {
	if !strings.Contains(a.serverMechanisms, saslAuthenticationProtocol) {
		return fmt.Errorf("not supported %s", saslAuthenticationProtocol)
	}

	buf := make([]byte, 18)
	_, err := rand.Read(buf)
	if err != nil {
		return err
	}

	a.clientFirstMessageBare = make([]byte, base64.RawStdEncoding.EncodedLen(len(buf)))
	base64.RawStdEncoding.Encode(a.clientFirstMessageBare, buf)

	return nil
}

func (s *scramAuth) receiveSASLContinue() (*SASLContinue, error) {
	msg, err := s.reader.Receive()
	if err != nil {
		return nil, err
	}

	if sasl, ok := msg.(*SASLContinue); ok {
		return sasl, nil
	}

	return nil, errors.New("expected sasl continue")
}

func (a *scramAuth) clientFinalMessage() string {
	clientFinalMessageWithoutProof := []byte(fmt.Sprintf("c=biws,r=%s", a.serverFirstMessage.r))

	a.saltedPassword = pbkdf2.Key(a.password, a.serverFirstMessage.s, a.serverFirstMessage.i, 32, sha256.New)
	a.authMessage = bytes.Join([][]byte{a.clientFirstMessage, a.serverFirstMessage.all, clientFinalMessageWithoutProof}, []byte(","))

	clientProof := computeClientProof(a.saltedPassword, a.authMessage)

	return fmt.Sprintf("%s,p=%s", clientFinalMessageWithoutProof, clientProof)
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

func (s *SASLContinue) isMessage() {}

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

	s.s, err = base64.StdEncoding.DecodeString(string(s.s))
	if err != nil {
		return nil, err
	}

	return s, nil
}

type SASLResponse struct {
	Data []byte
}

func (s SASLResponse) Encode() []byte {
	buf := make([]byte, 5)
	buf[0] = 'p'

	binary.BigEndian.PutUint32(buf[1:5], uint32(len(s.Data)+4))
	buf = append(buf, s.Data...)

	return buf
}

type SASLFinal struct {
	data []byte
}

func (s SASLFinal) isMessage() {}

func NewSASLFinal(data []byte) (SASLFinal, error) {
	return SASLFinal{}, nil
}

func (s *scramAuth) validateSASLFinal() error {
	return nil
}

func computeHMAC(key, msg []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	return mac.Sum(nil)
}

func computeClientProof(saltedPassword, authMessage []byte) []byte {
	clientKey := computeHMAC(saltedPassword, []byte("Client Key"))
	storedKey := sha256.Sum256(clientKey)
	clientSignature := computeHMAC(storedKey[:], authMessage)

	clientProof := make([]byte, len(clientSignature))
	for i := 0; i < len(clientSignature); i++ {
		clientProof[i] = clientKey[i] ^ clientSignature[i]
	}

	buf := make([]byte, base64.StdEncoding.EncodedLen(len(clientProof)))
	base64.StdEncoding.Encode(buf, clientProof)
	return buf
}*/
