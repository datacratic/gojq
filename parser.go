// Copyright (c) 2015 Datacratic. All rights reserved.

package jq

import (
	"errors"
	"io"
)

var (
	ErrArrayEndOrNext     = errors.New("expected array element or end")
	ErrFieldEndOrNext     = errors.New("expected field or object end")
	ErrExpectedFieldValue = errors.New("expected field value after name")
	ErrExpectedFalse      = errors.New("expected 'false'")
	ErrExpectedTrue       = errors.New("expected 'true'")
	ErrExpectedNull       = errors.New("expected 'null'")
)

type node struct {
	kind        int
	fieldOffset int
	fieldLength int
	valueOffset int
	valueLength int
	count       int
	up          int
	down        int
	next        int
}

type parser struct {
	bytes []byte
	nodes []node
	i     int
}

func (p *parser) parse(value *Value) (err error) {
	n := len(value.bytes)
	p.bytes = append(value.bytes, ']', '}', '"', ' ')
	p.nodes = append(value.nodes[:0], node{kind: TypeRoot})
	p.nodes[0].kind = TypeRoot
	value.bytes = p.bytes[:n]
	value.nodes = p.nodes[:0]
	if err = p.parseValue(0); err != nil {
		return
	}

	if p.i > n {
		err = io.EOF
		return
	}

	value.nodes = p.nodes
	return
}

func (p *parser) parseArray(item int) (err error) {
	p.nodes[item].kind = TypeArray

	if p.bytes[p.i] == ']' {
		p.i++
		return
	}

	p.nodes[item].down = len(p.nodes)
	for {
		index := len(p.nodes)
		p.nodes = append(p.nodes, node{up: item})
		p.nodes[item].count++
		if err = p.parseValue(index); err != nil {
			return
		}

		if p.bytes[p.i] == ',' {
			p.i++
			p.nodes[index].next = len(p.nodes)
			continue
		}

		if p.bytes[p.i] == ']' {
			p.i++
			break
		}

		err = ErrArrayEndOrNext
		break
	}

	return
}

func (p *parser) parseObject(item int) (err error) {
	p.nodes[item].kind = TypeObject

	if p.bytes[p.i] == '}' {
		p.i++
		return
	}

	p.nodes[item].down = len(p.nodes)
	for {
		if p.bytes[p.i] == '"' {
			p.i++

			index := len(p.nodes)
			p.nodes = append(p.nodes, node{up: item})
			p.nodes[item].count++
			if err = p.parseField(index); err != nil {
				return
			}

			if p.bytes[p.i] == ',' {
				p.i++
				p.nodes[index].next = len(p.nodes)
				continue
			}

			if p.bytes[p.i] == '}' {
				p.i++
				break
			}
		}

		err = ErrFieldEndOrNext
		break
	}

	return
}

func (p *parser) parseField(item int) (err error) {
	j := p.i

	for {
		if p.bytes[p.i] == '"' {
			p.i++
			break
		}

		if p.bytes[p.i] == '\\' {
			p.i++
		}

		p.i++
	}

	p.nodes[item].fieldOffset = j
	p.nodes[item].fieldLength = p.i - j

	if p.bytes[p.i] != ':' {
		err = ErrExpectedFieldValue
		return
	}

	p.i++
	err = p.parseValue(item)
	return
}

func (p *parser) parseString(item int) (err error) {
	p.nodes[item].kind = TypeString

	j := p.i

	for {
		if p.bytes[p.i] == '"' {
			p.i++
			break
		}

		if p.bytes[p.i] == '\\' {
			p.i++
		}

		p.i++
	}

	p.nodes[item].valueOffset = j
	p.nodes[item].valueLength = p.i - j
	return
}

func (p *parser) parseValue(item int) (err error) {
	switch p.bytes[p.i] {
	case '"':
		p.i++
		err = p.parseString(item)
		return
	case '{':
		p.i++
		err = p.parseObject(item)
		return
	case '[':
		p.i++
		err = p.parseArray(item)
		return
	}

	j := p.i

	switch p.bytes[p.i] {
	case 't':
		if p.bytes[p.i+1] != 'r' || p.bytes[p.i+2] != 'u' || p.bytes[p.i+3] != 'e' {
			err = ErrExpectedTrue
			return
		}

		p.i += 4
		p.nodes[item].kind = TypeTrue
	case 'f':
		if p.bytes[p.i+1] != 'a' || p.bytes[p.i+2] != 'l' || p.bytes[p.i+3] != 's' || p.bytes[p.i+4] != 'e' {
			err = ErrExpectedFalse
			return
		}

		p.i += 5
		p.nodes[item].kind = TypeFalse
	case 'n':
		if p.bytes[p.i+1] != 'u' || p.bytes[p.i+2] != 'l' || p.bytes[p.i+3] != 'l' {
			err = ErrExpectedNull
			return
		}

		p.i += 4
		p.nodes[item].kind = TypeNull
	default:
		for {
			switch p.bytes[p.i] {
			case ',', '}', ']':
			default:
				p.i++
				continue
			}

			break
		}

		p.nodes[item].kind = TypeNumber
	}

	p.nodes[item].valueOffset = j
	p.nodes[item].valueLength = p.i - j
	return
}
