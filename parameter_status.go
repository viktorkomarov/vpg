package vpg

import (
	"bytes"
)

type parameterStatus struct {
	Name  string
	Value string
}

func (p parameterStatus) isMessage() {}

func newParametrStatus(payload []byte) parameterStatus {
	params := bytes.Split(payload, []byte(" "))

	var value string
	if len(params) > 1 {
		value = string(params[1])
	}

	return parameterStatus{
		Name:  string(params[0]),
		Value: value,
	}
}
