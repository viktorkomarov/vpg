package encoder

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"sort"
)

func Encode(v interface{}) ([]byte, error) {
	fields, err := analyzeFields(v)
	if err != nil {
		return nil, err
	}

	sort.Slice(fields, func(i, j int) bool {
		return fields[i].order < fields[j].order
	})

	if err := checkSortedOrder(fields); err != nil {
		return nil, fmt.Errorf("%w: invalid bytes boundaries", err)
	}

	var out bytes.Buffer
	start := buildHeaders(fields[0], &out)
	if start == 1 {
		fields = fields[1:]
	}

	for _, field := range fields {
		encode, err := encoderByType(field.val)
		if err != nil {
			return nil, err
		}

		if field.preffix != "" {
			err = encodeString(&out, reflect.ValueOf(field.preffix))
			if err != nil {
				return nil, err
			}
		}

		if err = encode(&out, field.val); err != nil {
			return nil, err
		}
	}

	data := out.Bytes()
	binary.BigEndian.PutUint32(data[start:start+4], uint32(len(data)))

	return data, nil
}

func buildHeaders(field Field, out *bytes.Buffer) int {
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
		return encodeSlice, nil
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
		elem := v.Index(i)
		encoder, err := encoderByType(elem)
		if err != nil {
			return err
		}

		err = encoder(out, elem)
		if err != nil {
			return err
		}
	}

	return nil
}
