package protocol

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type RESPType byte

const (
	SimpleString RESPType = '+'
	ErrorType    RESPType = '-'
	Integer      RESPType = ':'
	BulkString   RESPType = '$'
	Array        RESPType = '*'
)

// Value RESP 值。
type Value struct {
	Type  RESPType
	Str   string
	Num   int64
	Array []Value
}

// Parse 解析 RESP。
func Parse(reader *bufio.Reader) (*Value, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch RESPType(prefix) {
	case SimpleString, ErrorType:
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			return nil, readErr
		}
		return &Value{Type: RESPType(prefix), Str: strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")}, nil
	case Integer:
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			return nil, readErr
		}
		n, parseErr := strconv.ParseInt(strings.TrimSpace(line), 10, 64)
		if parseErr != nil {
			return nil, parseErr
		}
		return &Value{Type: Integer, Num: n}, nil
	case BulkString:
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			return nil, readErr
		}
		length, parseErr := strconv.Atoi(strings.TrimSpace(line))
		if parseErr != nil {
			return nil, parseErr
		}
		if length < 0 {
			return &Value{Type: BulkString, Str: ""}, nil
		}
		buf := make([]byte, length+2)
		if _, readErr = reader.Read(buf); readErr != nil {
			return nil, readErr
		}
		return &Value{Type: BulkString, Str: string(buf[:length])}, nil
	case Array:
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			return nil, readErr
		}
		count, parseErr := strconv.Atoi(strings.TrimSpace(line))
		if parseErr != nil {
			return nil, parseErr
		}
		items := make([]Value, 0, count)
		for i := 0; i < count; i++ {
			item, itemErr := Parse(reader)
			if itemErr != nil {
				return nil, itemErr
			}
			items = append(items, *item)
		}
		return &Value{Type: Array, Array: items}, nil
	default:
		return nil, errors.New("unsupported resp type")
	}
}

// Serialize 序列化 RESP。
func Serialize(value *Value) []byte {
	if value == nil {
		return []byte("$-1\r\n")
	}

	switch value.Type {
	case SimpleString:
		return []byte("+" + value.Str + "\r\n")
	case ErrorType:
		return []byte("-" + value.Str + "\r\n")
	case Integer:
		return []byte(fmt.Sprintf(":%d\r\n", value.Num))
	case BulkString:
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value.Str), value.Str))
	case Array:
		buf := bytes.NewBufferString(fmt.Sprintf("*%d\r\n", len(value.Array)))
		for i := range value.Array {
			item := value.Array[i]
			buf.Write(Serialize(&item))
		}
		return buf.Bytes()
	default:
		return []byte("-ERR unsupported type\r\n")
	}
}
