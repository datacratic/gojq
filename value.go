// Copyright (c) 2015 Datacratic. All rights reserved.

package jq

import (
	"bytes"
	"io"
	"strconv"
	"unsafe"
)

type Kind int

const (
	Root Kind = iota
	Array
	Object
	String
	Number
	True
	False
	Null
	Unknown
)

func (k Kind) String() string {
	switch k {
	case Array:
		return "Array"
	case Object:
		return "Object"
	case String:
		return "String"
	case Number:
		return "Number"
	case True:
		return "True"
	case False:
		return "False"
	case Null:
		return "Null"
	default:
		return "Unknown"
	}
}

type Value struct {
	nodes []node
	bytes []byte
}

func (value *Value) Unmarshal(data []byte) (err error) {
	value.bytes = data
	p := parser{}
	err = p.parse(value)
	return
}

func (value *Value) UnmarshalFrom(r io.Reader) (err error) {
	b := bytes.NewBuffer(value.bytes[:0])

	if _, err = b.ReadFrom(r); err != nil {
		return
	}

	value.bytes = b.Bytes()
	p := parser{}
	err = p.parse(value)
	return
}

func (value *Value) NewQuery() Query {
	return Query{
		value: value,
		index: 0,
	}
}

func (value *Value) Extract(keys ...string) (result interface{}) {
	q := value.NewQuery()

	i, err := q.find(keys)
	if err != nil {
		return
	}

	switch value.nodes[i].kind {
	case String:
		offset := value.nodes[i].valueOffset
		length := value.nodes[i].valueLength
		s := string(value.bytes[offset : offset+length])
		result = s
	case Number:
		offset := value.nodes[i].valueOffset
		length := value.nodes[i].valueLength
		s := string(value.bytes[offset : offset+length])
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return
		}

		result = f
	case True:
		result = true
	case False:
		result = false
	case Null:
		result = nil
	}

	return
}

func (value *Value) findFrom(index int, key string) int {
	i := value.nodes[index].down

	// see http://stackoverflow.com/a/31484416/2611792
	k := *(*[]byte)(unsafe.Pointer(&key))
	n := len(k)

	for {
		length := value.nodes[i].fieldLength
		if n == length {
			offset := value.nodes[i].fieldOffset
			if bytes.Equal(k, value.bytes[offset:offset+length]) {
				return i
			}
		}

		if i = value.nodes[i].next; i == 0 {
			break
		}
	}

	return -1
}
