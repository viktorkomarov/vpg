package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
)

type test struct {
	H byte   `header:"p"`
	A int    `pg_type:"1" pg_name:"user"`
	B string `pg_type:"2"`
}

type DecodeMe struct {
	H       byte   `header:"R"`
	Payload string `pg_type:"1"`
}

func main() {
	var d DecodeMe
	d.Payload = "Payload"
	data, err := Encode(d)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%v\n", data)

	var b DecodeMe
	err = Decode(data, &b)
	if err != nil {
		log.Fatal(err)
	}
}

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

	header := headerSize(data)
	data = data[5:] // skip header and size

	if header != fields[0].header {
		// handler postgres err and other async msg
		return errors.New("mismatch header type")
	}

	fields = fields[:1]
	for _, field := range fields {
		fmt.Println(field.val.Kind())
		decode, err := decodeByType(field.val)
		if err != nil {
			return err
		}

		offset := lenOfVal(data, field.val)
		if offset > len(data) {
			return errors.New("real problem")
		}

		if err = decode(data[:offset], field.val); err != nil {
			return err
		}
	}

	return nil
}

func headerSize(data []byte) byte {
	header := data[0]
	//size := binary.BigEndian.Uint32(data[1:5])

	return header
}

func decodeByType(v reflect.Value) (func([]byte, reflect.Value) error, error) {
	switch v.Kind() {
	case reflect.String:
		return decodeString, nil
	default:
		return nil, fmt.Errorf("unsupported decode type %+v", v.Kind())
	}
}

func decodeString(b []byte, v reflect.Value) error {
	if !v.CanSet() {
		v.SetString(string(b))
	}

	return fmt.Errorf("unsettable field %+v", v)
}

func lenOfVal(data []byte, v reflect.Value) int {
	switch v.Kind() {
	case reflect.String:
		return bytes.IndexByte(data, '\000')
	case reflect.Int16:
		return 2
	case reflect.Int32:
		return 4
	case reflect.Uint8:
		return 1
	case reflect.Slice: // get size by prev field or by default value create own Type like alloc
		return 0
	}

	return 0
}
