package common

import (
	"strconv"
	"strings"
)

// General buffer unit design
/*
 * CmBuf:     The basic interface for buffer
 * simpleBuf: The simplest buffer. However, it assumes all data are integers
 */

type CmBuf interface {
	GetBuf() []byte
	ResetBuf(buf []byte)
	ToString() string                                         // Return data as byte string
	GetFieldAsString() string                                 // Return name of the fields as byte string
	SetField(name string, datum interface{}, jsonIgnore bool) // Set datum to buffer. If jsonIgnore is true, the field would not appear in toString
	GetField(name string) (interface{}, bool)                 // Return data corresponding to name and whether data can be found
}

type simpleBuf struct {
	dataKey    []string
	dataVal    []interface{}
	jsonIgnore []bool
	buf        []byte
}

func (b *simpleBuf) GetBuf() []byte {
	return b.buf
}

func (b *simpleBuf) ResetBuf(buf []byte) {
	b.buf = buf
}

func (b *simpleBuf) ToString() string {
	valArr := make([]string, 0)
	for idx, datum := range b.dataVal {
		if b.jsonIgnore[idx] {
			continue
		}
		if val, isInt := datum.(int); isInt {
			valArr = append(valArr, strconv.Itoa(val))
			continue
		}
		if val, isStr := datum.(string); isStr {
			valArr = append(valArr, val)
			continue
		}
	}
	rv := strings.Join(valArr, ",")
	rv += "\n"
	return rv
}

func (b *simpleBuf) GetFieldAsString() string {
	keyArr := make([]string, 0)
	for idx, k := range b.dataKey {
		if b.jsonIgnore[idx] {
			continue
		}
		keyArr = append(keyArr, k)
	}
	rv := strings.Join(keyArr, ",")
	rv += "\n"
	return rv
}

// Add a field entry to buffer.
// Adding an existing field overwrites the value
func (b *simpleBuf) SetField(name string, datum interface{}, jsonIgnore bool) {
	fieldIdx := -1
	for idx, f := range b.dataKey {
		if f == name {
			fieldIdx = idx
			break
		}
	}
	if fieldIdx == -1 {
		b.dataKey = append(b.dataKey, name)
		b.dataVal = append(b.dataVal, datum)
		b.jsonIgnore = append(b.jsonIgnore, jsonIgnore)
	} else {
		b.dataVal[fieldIdx] = datum
		b.jsonIgnore[fieldIdx] = jsonIgnore
	}
}

func (b *simpleBuf) GetField(name string) (interface{}, bool) {
	for i := range b.dataKey {
		if b.dataKey[i] == name {
			return b.dataVal[i], true
		}
	}
	return nil, false
}

func GetBufFieldAsInt(b CmBuf, name string) (int, bool) {
	field, ok := b.GetField(name)
	rv, isInt := field.(int)
	return rv, ok && isInt
}

func GetBufFieldAsString(b CmBuf, name string) (string, bool) {
	field, ok := b.GetField(name)
	rv, isString := field.(string)
	return rv, ok && isString
}

func MakeSimpleBuf(inBuf []byte) *simpleBuf {
	rv := simpleBuf{dataKey: make([]string, 0), dataVal: make([]interface{}, 0), jsonIgnore: make([]bool, 0), buf: inBuf}
	return &rv
}
