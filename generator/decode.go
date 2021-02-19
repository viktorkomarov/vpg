package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
)

func Decode(data []byte, v interface{}) error {
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Ptr || ptr.IsNil() {
		return errors.New("not nil ptr is required")
	}

	fields, err := analyzeFields(v)
	if err != nil {
		return err
	}

	sort.Slice(fields, func(i, j int) bool {
		return fields[i].order < fields[j].order
	})

	header := data[0]
	data = data[5:] // skip header and size

	if header != fields[0].header {
		// handler postgres err and other async msg
		return errors.New("mismatch header type")
	}

	fields = fields[1:]
	for _, field := range fields {
		decode, err := decodeByType(field.typ)
		if err != nil {
			return err
		}

		offset := lenOfVal(data, field.val)
		if offset > len(data) {
			return errors.New("real problem")
		}

		log.Printf("%+v\n", data[:offset])
		if err = decode(data[:offset], field.val); err != nil {
			return err
		}

		data = data[offset:]
	}

	return nil
}

func decodeByType(v reflect.Type) (func([]byte, reflect.Value) error, error) {
	switch v.Kind() {
	case reflect.String:
		return decodeString, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return decodeInt, nil
	default:
		return nil, fmt.Errorf("unsupported decode type %+v", v.Kind())
	}
}

func decodeString(b []byte, v reflect.Value) error {
	if v.CanSet() {
		v.SetString(string(b))
		return nil
	}

	return fmt.Errorf("unsettable field %+v", v)
}

func decodeInt(b []byte, v reflect.Value) error {
	if v.CanSet() {
		data := int64(binary.BigEndian.Uint32(b))
		v.SetInt(data)
		return nil
	}

	return fmt.Errorf("unsettable field %+v", v)
}

func lenOfVal(data []byte, v reflect.Value) int {
	switch v.Kind() {
	case reflect.String:
		return bytes.IndexByte(data, '\000') + 1
	case reflect.Int16:
		return 2
	case reflect.Int32, reflect.Int64, reflect.Int:
		return 4
	case reflect.Uint8:
		return 1
	case reflect.Slice: // -1 if size prev
		return 0
	}

	return 0
}
