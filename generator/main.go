package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"sort"
	"strconv"
)

type test struct {
	H byte   `header:"p"`
	A int    `pg_type:"1" pg_name:"user"`
	B string `pg_type:"2"`
}

func main() {

}

type BytesFields struct {
	header byte
	order  int
	name   string
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

		if bytesMap[i].name != "" {
			err = encodeString(&out, reflect.ValueOf(bytesMap[i].name))
			if err != nil {
				return nil, err
			}
		}

		if err = encoder(&out, bytesMap[i].val); err != nil {
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

		pg, ok := field.Tag.Lookup("pg_type")
		if !ok {
			continue
		}

		name, _ := field.Tag.Lookup("pg_name")
		order, err := strconv.Atoi(pg)
		if err != nil {
			return nil, fmt.Errorf("incorrect order %s", pg)
		}

		bytesMap = append(bytesMap, BytesFields{
			order: order,
			name:  name,
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

func encoderByType(v reflect.Value) (func(*bytes.Buffer, reflect.Value) error, error) {
	switch v.Kind() {
	case reflect.Uint8: // byte is alias for uint8
		return encodeByte, nil
	case reflect.Int16:
		return encodeInt16, nil
	case reflect.Int, reflect.Int32, reflect.Int64:
		return encodeInt, nil
	case reflect.String:
		return encodeString, nil
	case reflect.Slice:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported pg type %v", v)
	}
}

func encodeByte(out *bytes.Buffer, v reflect.Value) error {
	return out.WriteByte(v.Interface().(byte))
}

func encodeInt16(out *bytes.Buffer, v reflect.Value) error {
	out.Grow(2)
	return binary.Write(out, binary.BigEndian, v.Interface().(int16))
}

func encodeInt(out *bytes.Buffer, v reflect.Value) error {
	out.Grow(4)
	return binary.Write(out, binary.BigEndian, int32(v.Int()))
}

func encodeString(out *bytes.Buffer, v reflect.Value) error {
	_, err := out.Write(append([]byte(v.Interface().(string)), '\000'))
	return err
}

func encodeSlice(out *bytes.Buffer, v reflect.Value) error {
	for i := 0; i < v.Len(); i++ {
		encoder, err := encoderByType(v.Index(i))
		if err != nil {
			return err
		}

		err = encoder(out, v)
		if err != nil {
			return err
		}
	}

	return nil
}
