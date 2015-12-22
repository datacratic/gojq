// Copyright (c) 2015 Datacratic. All rights reserved.

package jq

import (
	"fmt"
	"strconv"
	"strings"
)

type Query struct {
	value *Value
	index int
}

func (q *Query) Array(keys ...string) (err error) {
	i, err := q.getIndex(TypeArray, keys)
	if err != nil {
		return
	}

	q.index = i
	return
}

func (q *Query) Object(keys ...string) (err error) {
	i, err := q.getIndex(TypeObject, keys)
	if err != nil {
		return
	}

	q.index = i
	return
}

func (q Query) ForEach(callback func(q Query) error) (err error) {
	kind := q.value.nodes[q.index].kind

	if kind != TypeObject && kind != TypeArray {
		err = fmt.Errorf("node is not an object or array")
		return
	}

	index := q.value.nodes[q.index].down
	if index == 0 {
		return
	}

	for {
		err = callback(Query{value: q.value, index: index})
		if err != nil {
			break
		}

		if index = q.value.nodes[index].next; index == 0 {
			break
		}
	}

	return
}

func (q Query) String(keys ...string) (result string, err error) {
	i, err := q.getIndex(TypeString, keys)
	if err != nil {
		return
	}

	offset := q.value.nodes[i].valueOffset
	length := q.value.nodes[i].valueLength - 1
	result = string(q.value.bytes[offset : offset+length])
	return
}

func (q Query) Float64(keys ...string) (result float64, err error) {
	i, err := q.getIndex(TypeNumber, keys)
	if err != nil {
		return
	}

	offset := q.value.nodes[i].valueOffset
	length := q.value.nodes[i].valueLength
	s := string(q.value.bytes[offset : offset+length])
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return
	}

	result = f
	return
}

func (q Query) Int64(keys ...string) (result int64, err error) {
	f, err := q.Float64(keys...)
	if err != nil {
		return
	}

	result = int64(f)
	if float64(result) != f {
		err = fmt.Errorf("number '%f' is not an integer", f)
	}

	return
}

func (q Query) findIndex(kind int, keys []string) (index int) {
	index = q.index

	for _, key := range keys {
		index = q.value.findFrom(index, key)
		if index < 0 {
			return
		}
	}

	if kind != q.value.nodes[index].kind {
		index = ^index
	}

	return
}

func (q Query) getIndex(kind int, keys []string) (result int, err error) {
	index := q.index

	for i, key := range keys {
		index = q.value.findFrom(index, key)
		if index < 0 {
			err = fmt.Errorf("key '%s' is missing at '%s'", key, strings.Join(keys[:i], "."))
			return
		}
	}

	if kind != q.value.nodes[index].kind {
		err = fmt.Errorf("key '%s' type is unexpected", strings.Join(keys, "."))
	} else {
		result = index
	}

	return
}
