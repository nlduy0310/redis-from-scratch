package server

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	SIMPLE_STRING = "+"
	ERROR         = "-"
	INTEGER       = ":"
	BULK_STRING   = "$"
	ARRAY         = "*"
)

type RespValue struct {
	vtype   string
	str     string
	err     string
	num     int
	bulkStr string
	array   []RespValue
}

func (v RespValue) String() string {

	switch v.vtype {
	case SIMPLE_STRING:
		return fmt.Sprintf("{SIMPLE_STRING: %s}", v.str)
	case ERROR:
		return fmt.Sprintf("{ERROR: %s}", v.err)
	case INTEGER:
		return fmt.Sprintf("{INTEGER: %d}", v.num)
	case BULK_STRING:
		return fmt.Sprintf("{BULK_STRING: %s}", v.bulkStr)
	case ARRAY:
		valueStrs := make([]string, len(v.array))
		for i := range len(v.array) {
			valueStrs[i] = v.array[i].String()
		}
		return fmt.Sprintf("{ARRAY: [%s]}", strings.Join(valueStrs, ", "))
	default:
		panic(fmt.Sprintf("found invalid type when converting value to string: %s", v.vtype))
	}
}

func (v RespValue) Equal(other RespValue) bool {

	switch v.vtype {
	case SIMPLE_STRING:
		return v.vtype == other.vtype && v.str == other.str
	case ERROR:
		return v.vtype == other.vtype && v.err == other.err
	case INTEGER:
		return v.vtype == other.vtype && v.num == other.num
	case BULK_STRING:
		return v.vtype == other.vtype && v.bulkStr == other.bulkStr
	case ARRAY:
		{
			if v.vtype != other.vtype || len(v.array) != len(other.array) {
				return false
			}
			for i := range len(v.array) {
				if !v.array[i].Equal(other.array[i]) {
					return false
				}
			}
			return true
		}
	default:
		panic(fmt.Sprintf("found invalid type when converting value to string: %s", v.vtype))
	}
}

type Resp struct {
	reader *bufio.Reader
}

func NewResp(r io.Reader) *Resp {

	return &Resp{reader: bufio.NewReader(r)}
}

func (resp *Resp) readLine() (line []byte, dataBytesRead int, err error) {

	for {
		b, err := resp.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}

		dataBytesRead++
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' && line[len(line)-1] == '\n' {
			break
		}
	}
	return line[:len(line)-2], dataBytesRead - 2, nil
}

// assuming we're on a 64-bit system, then converting int64 to int is safe
func (resp *Resp) readInteger() (result int, dataBytesRead int, err error) {

	line, dataBytesRead, err := resp.readLine()
	if err != nil {
		return 0, 0, err
	}

	parsedInt, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return int(parsedInt), dataBytesRead, nil
}

func (resp *Resp) readBulkString() (result string, dataBytesRead int, err error) {

	length, _, err := resp.readInteger()
	if err != nil {
		return "", 0, err
	}

	line, bytes, err := resp.readLine()
	if err != nil {
		return "", 0, err
	}

	if len(line) != length {
		return "", 0, fmt.Errorf("invalid bulk string, length and data mismatch")
	}

	return string(line), bytes, nil
}

func (resp *Resp) readArray() (result []RespValue, err error) {

	length, _, err := resp.readInteger()
	if err != nil {
		return nil, err
	}

	result = make([]RespValue, length)
	for i := range result {
		v, err := resp.Read()
		if err != nil {
			return nil, fmt.Errorf("error reading element at index %d: %s", i, err.Error())
		}
		result[i] = v
	}

	return result, nil
}

func (resp *Resp) Read() (RespValue, error) {

	b, err := resp.reader.ReadByte()

	if err != nil {
		return RespValue{}, err
	}

	switch string(b) {
	case SIMPLE_STRING:
		{
			line, _, err := resp.readLine()
			if err != nil {
				return RespValue{}, fmt.Errorf("error while reading simple string: %s", err.Error())
			}
			return RespValue{vtype: SIMPLE_STRING, str: string(line)}, nil
		}
	case ERROR:
		{
			line, _, err := resp.readLine()
			if err != nil {
				return RespValue{}, fmt.Errorf("error while reading simple error: %s", err.Error())
			}
			return RespValue{vtype: ERROR, err: string(line)}, nil
		}
	case INTEGER:
		{
			num, _, err := resp.readInteger()
			if err != nil {
				return RespValue{}, fmt.Errorf("error while reading simple integer: %s", err.Error())
			}
			return RespValue{vtype: INTEGER, num: num}, nil
		}
	case BULK_STRING:
		{
			line, _, err := resp.readBulkString()
			if err != nil {
				return RespValue{}, fmt.Errorf("error while reading bulk string: %s", err.Error())
			}
			return RespValue{vtype: BULK_STRING, bulkStr: line}, nil
		}
	case ARRAY:
		{
			values, err := resp.readArray()
			if err != nil {
				return RespValue{}, fmt.Errorf("error while reading array: %s", err.Error())
			}
			return RespValue{vtype: ARRAY, array: values}, nil
		}
	default:
		return RespValue{}, fmt.Errorf("unknown data type '%s'", string(b))
	}
}
