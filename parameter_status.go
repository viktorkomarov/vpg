package main

import (
	"bytes"
)

type ParameterStatus struct {
	Name  string
	Value string
}

func (p *ParameterStatus) isMessage() {}

func NewParametrStatus(payload []byte) *ParameterStatus {
	params := bytes.Split(payload, []byte(" "))
	if len(params) < 2 { // doesn't matter
		return &ParameterStatus{
			Name:  string(params[0]),
			Value: "",
		}
	}

	return &ParameterStatus{
		Name:  string(params[0]),
		Value: string(params[1]),
	}
}
