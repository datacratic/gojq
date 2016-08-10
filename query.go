// Copyright (c) 2015 Datacratic. All rights reserved.

package jq

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrNotFound = errors.New("not found")
	ErrEnd      = errors.New("end")
)

type Query struct {
	value *Value
	index int
}

func (q *Query) Key() (result string) {
	length := q.value.nodes[q.index].fieldLength
	if length == 0 {
		return
	}

	offset := q.value.nodes[q.index].fieldOffset
	result = string(q.value.bytes[offset : offset+length])
	return
}

func (q *Query) Value() (result string) {
	length := q.value.nodes[q.index].valueLength
	if length == 0 {
		return
	}

	offset := q.value.nodes[q.index].valueOffset
	result = string(q.value.bytes[offset : offset+length])
	return
}

func (q *Query) Kind() Kind {
	return q.value.nodes[q.index].kind
}

func (q *Query) Next() bool {
	i := q.value.nodes[q.index].next
	if i == 0 {
		return false
	}

	q.index = i
	return true
}

func (q *Query) At(n int) bool {
	i := q.value.nodes[q.index].down
	if i == 0 {
		return false
	}

	for k := 0; k < n; k++ {
		if i = q.value.nodes[i].next; i == 0 {
			return false
		}
	}

	q.index = i
	return true
}

func (q *Query) AtKey(key, value string) bool {
	w := q.index
	i := q.value.nodes[q.index].down

	for i != 0 {
		j := q.value.nodes[i].down
		for j != 0 {
			q.index = j
			if q.Key() == key && q.Value() == value {
				q.index = i
				return true
			}

			j = q.value.nodes[j].next
		}

		i = q.value.nodes[i].next
	}

	q.index = w
	return false
}

func (q *Query) Back() bool {
	i := q.value.nodes[q.index].back
	if i == 0 {
		return false
	}

	q.index = i
	return true
}

func (q *Query) Count() int {
	return q.value.nodes[q.index].count
}

func (q *Query) Down() bool {
	i := q.value.nodes[q.index].down
	if i == 0 {
		return false
	}

	q.index = i
	return true
}

func (q *Query) Up() bool {
	i := q.value.nodes[q.index].up
	if i == 0 {
		return false
	}

	q.index = i
	return true
}

func (q *Query) FindArray(keys ...string) (err error) {
	i, err := q.getIndex(Array, keys)
	if err != nil {
		return
	}

	q.index = i
	return
}

func (q *Query) FindObject(keys ...string) (err error) {
	i, err := q.getIndex(Object, keys)
	if err != nil {
		return
	}

	q.index = i
	return
}

func (q *Query) Walk(callback func(*Query) error) (err error) {
	w := q.index
	i := w

	for {
		if err = callback(q); err != nil {
			break
		}

		j := q.value.nodes[i].down
		if j != 0 {
			q.index = j
			if err = q.Walk(callback); err != nil {
				break
			}
		}

		if i = q.value.nodes[i].next; i == 0 {
			break
		}

		q.index = i
	}

	q.index = w
	return
}

func (q *Query) String(keys ...string) (result string, err error) {
	i, err := q.getIndex(String, keys)
	if err != nil {
		return
	}

	offset := q.value.nodes[i].valueOffset
	length := q.value.nodes[i].valueLength
	result = string(q.value.bytes[offset : offset+length])
	return
}

func (q *Query) Float64(keys ...string) (result float64, err error) {
	i, err := q.getIndex(Number, keys)
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

func (q *Query) Int64(keys ...string) (result int64, err error) {
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

func (q *Query) find(keys []string) (result int, err error) {
	index := q.index

	for i, n := 0, len(keys); i < n; i++ {
		key := keys[i]

		switch key[0] {
		case '@':
			j := 0
			if j, err = strconv.Atoi(key[1:]); err != nil {
				return
			}

			p := &Query{value: q.value, index: index}
			if !p.At(j) {
				err = ErrNotFound
				return
			}

			index = p.index
		case '$':
			p := &Query{value: q.value, index: index}
			if !p.AtKey(key[1:], keys[i+1]) {
				err = ErrNotFound
				return
			}

			index = p.index
			i++
		default:
			if index = q.value.findFrom(index, key); index < 0 {
				err = ErrNotFound
				return
			}
		}
	}

	result = index
	return
}

func (q *Query) getIndex(kind Kind, keys []string) (result int, err error) {
	index, err := q.find(keys)
	if err != nil {
		return
	}

	if k := q.value.nodes[index].kind; k != kind {
		err = fmt.Errorf("expecting type '%s' instead of '%s'", kind, k)
	} else {
		result = index
	}

	return
}
