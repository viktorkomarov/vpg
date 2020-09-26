package main

// https://postgrespro.ru/docs/postgrespro/10/protocol-error-fields

import (
	"bytes"
	"fmt"
)

type ErrPostgresResponse struct {
	level    string
	level2   string // always equal level but not translated
	sqlState string
	hMsg     string
}

func NewErrPostgresResponse(data []byte) error {
	forms := bytes.Split(data, []byte{'\000'})
	err := &ErrPostgresResponse{}

	for _, form := range forms {
		if len(form) < 1 {
			continue
		}

		errType := rune(form[0])
		switch errType {
		case 'S':
			err.level = string(form[1:])
		case 'V':
			err.level2 = string(form[1:])
		case 'C':
			err.sqlState = string(form[1:])
		case 'M':
			err.hMsg = string(form[1:])
		}
	}

	return err
}

func (e *ErrPostgresResponse) Error() string {
	return fmt.Sprintf("level: %s code : %s msg : %s", e.level2, e.sqlState, e.hMsg)
}
