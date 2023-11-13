package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleBufOverwriteField(t *testing.T) {
	buf := MakeSimpleBuf([]byte{})
	field := "dummy"

	buf.SetField(field, 0, true)

	v1, ok := buf.GetField(field)
	if !ok {
		panic("No such field")
	}
	assert.Equal(t, 0, v1, "Field should be 0")

	buf.SetField(field, 10, false)

	v2, ok := buf.GetField(field)
	if !ok {
		panic("No such field")
	}
	assert.Equal(t, 10, v2, "Field should be 10")
}

func TestIOUnitWithSimpleBuf(t *testing.T) {
	buf := []byte{1, 2, 3}
	buffer := MakeSimpleBuf(buf)

	buffer.SetField("dummy2", 50, true)

	unit := MakeIOUnit(buffer, -1, -1)

	// Test if *IOUnit implements CmUnit
	var _ CmUnit = (*IOUnit)(unit)

	extractedBuffer := unit.GetBuf()

	extractedBuffer.SetField("dummy", 100, false)

	if field, hasField := extractedBuffer.GetField("dummy"); hasField {
		if v, isInt := field.(int); isInt {
			assert.Equal(t, v, 100, "dummy field should be 100")
		} else {
			panic(fmt.Sprintf("Data not int but %T", field))
		}
	} else {
		panic("No data found")
	}

	assert.Equal(t, buf, extractedBuffer.GetBuf(), "buf should be [1, 2, 3]")
	assert.Equal(t, "dummy\n", extractedBuffer.GetFieldAsString(), "buf field should be dummy")
	assert.Equal(t, "100\n", extractedBuffer.ToString(), "buf string should be 100")

	_, hasField := extractedBuffer.GetField("hi")
	assert.Equal(t, false, hasField, "buf does not have field hi")
}

func TestReadHex(t *testing.T) {
	r := GetBufferReader([]byte{0x45, 0x4E, 0x47})
	assert.Equal(t, 3, r.GetSize(), "Reader buffer size incorrect")
	assert.Equal(t, "45 4e 47", r.ReadHex(3), "Output not equal")
}

func TestReadChar(t *testing.T) {
	r := GetBufferReader([]byte{0x45, 0x4E, 0x47})
	assert.Equal(t, "ENG", r.ReadChar(3), "Output not equal")
}

func TestReadPcr(t *testing.T) {
	buf := []byte{0x0e, 0x26, 0xe0, 0x33, 0x7e, 0x11}
	r := GetBufferReader(buf)

	val := r.ReadBits(33)
	assert.Equal(t, val, 474857574, "value should be 474857574")
	
	assert.Equal(t, 4, r.GetPos(), "Reader is not reading at desired position")
	assert.Equal(t, 7, r.GetOffset(), "Reader is not reading at desired offset")
	assert.Equal(t, 2, len(r.GetRemainedBuffer()), "Reader remained buffer size incorrect")
	
	val = r.ReadBits(6)
	assert.Equal(t, val, 63, "value should be 63")
	
	val = r.ReadBits(9)
	assert.Equal(t, val, 17, "value should be 17")
}

func TestReadExpGolomb(t *testing.T) {
	r := GetBufferReader([]byte{0b00010000, 0b11001010})
	expected := []int{7, 2, 4}
	for _, exp := range expected {
		assert.Equal(t, exp, r.ReadExpGolomb())
	}
}

func TestBsWriter(t *testing.T) {
	writer := GetBufferWriter(9)

	writer.WriteByte(0x47)
	writer.Write(0, 1)
	writer.Write(0, 1)
	writer.Write(0, 1)
	writer.Write(33, 13)
	writer.WriteShort(500)
	writer.WriteInt(100000)

	expected := []byte{0x47, 0x00, 0x21, 0x01, 0xf4, 0x00, 0x01, 0x86, 0xa0}
	assert.Equal(t, expected, writer.GetBuf(), "Expected bytes: [0x47, 0x00, 0x21, 0x01, 0xf4, 0x00, 0x01, 0x86, 0xa0]")
}
