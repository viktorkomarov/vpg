package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
)

type test struct {
	H byte   `header:"p"`
	A int    `pg_type:"1"`
	B string `pg_type:"2"`
}

func main() {
	t := test{
		A: 5,
		B: "hello",
	}

	data, err := Encode(t)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", data)
}

type BytesFields struct {
	header byte
	order  int
	val    reflect.Value
}

func Decode(data []byte, v interface{}) error {
	return nil
}

func Encode(v interface{}) ([]byte, error) {
	bytesMap, err := analyzeFields(v)
	if err != nil {
		return nil, err
	}

	sort.Slice(bytesMap, func(i, j int) bool {
		return bytesMap[i].order < bytesMap[j].order
	})

	if err := checkBoundaries(bytesMap); err != nil {
		return nil, fmt.Errorf("%w: invalid bytes boundaries", err)
	}

	var out bytes.Buffer
	start := buildHeaders(bytesMap[0], &out)

	for i := 1; i < len(bytesMap); i++ {
		encoder, err := encoderByType(bytesMap[i].val)
		if err != nil {
			return nil, err
		}

		if err = encoder(&out); err != nil {
			return nil, err
		}
	}

	data := out.Bytes()
	binary.BigEndian.PutUint32(data[start:start+4], uint32(len(data)-1))

	return data, nil
}

func analyzeFields(v interface{}) ([]BytesFields, error) {
	vValue := reflect.Indirect(reflect.ValueOf(v))
	vType := reflect.TypeOf(v)

	if vValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("args isn't struct %+v", v)
	}

	bytesMap := make([]BytesFields, 0)
	for i := 0; i < vType.NumField(); i++ {
		field := vType.Field(i)
		header, ok := field.Tag.Lookup("header")
		if ok {
			bytesMap = append(bytesMap, BytesFields{
				order:  0,
				header: header[0],
			})

			continue
		}

		pg, ok := field.Tag.Lookup("pg")
		if !ok {
			continue
		}

		order, err := strconv.Atoi(pg)
		if err != nil {
			return nil, fmt.Errorf("incorrect order %s", pg)
		}

		bytesMap = append(bytesMap, BytesFields{
			order: order,
			val:   vValue.Field(i),
		})
	}

	return bytesMap, nil
}

func checkBoundaries(fields []BytesFields) error {
	uniq := make(map[int]bool)

	orders := 1
	for i := range fields {
		if fields[i].header != 0 {
			uniq[0] = true
			continue
		}

		if uniq[fields[i].order] {
			return fmt.Errorf("duplicate %d", fields[i].order)
		}

		if orders != fields[i].order {
			return fmt.Errorf("broken order %d", fields[i].order)
		}

		uniq[fields[i].order] = true
		orders++
	}

	return nil
}

func buildHeaders(field BytesFields, out *bytes.Buffer) int {
	begin := 0
	if field.header != 0 {
		out.WriteByte(field.header)
		begin++
	}

	out.Write([]byte{0, 0, 0, 0})
	return begin
}

func encoderByType(v reflect.Value) (func(out *bytes.Buffer) error, error) {
	switch v.Kind() {
	case reflect.Uint8: // byte is alias for uint8
		return func(out *bytes.Buffer) error {
			return out.WriteByte(v.Interface().(byte))
		}, nil
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(out *bytes.Buffer) error {
			return binary.Write(out, binary.BigEndian, v.Interface())
		}, nil
	case reflect.String:
		return func(out *bytes.Buffer) error {
			bytes := append([]byte(v.Interface().(string)), '\000')
			_, err := out.Write(bytes)
			return err
		}, nil
	case reflect.Array:
		return func(out *bytes.Buffer) error {
			return errors.New("impement me!")
		}, nil
	default:
		return nil, fmt.Errorf("unsupported pg type %v", v)
	}
}
