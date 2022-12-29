package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleBuf(t *testing.T) {
	buf := []byte{1, 2, 3}
	simpleBuf := MakeSimpleBuf(buf)
	simpleBuf.SetField("dummy", 100, false)

	if field, hasField := simpleBuf.GetField("dummy"); hasField {
		if v, isInt := field.(int); isInt {
			assert.Equal(t, v, 100, "dummy field should be 100")
		} else {
			panic("Data not int")
		}
	} else {
		panic("No data found")
	}

	assert.Equal(t, simpleBuf.GetFieldAsString(), "dummy\n", "Field should be dummy")
	assert.Equal(t, simpleBuf.ToString(), "100\n", "buf value should be 100")
}

func TestReadPcr(t *testing.T) {
	buf := []byte{0x0e, 0x26, 0xe0, 0x33, 0x7e, 0x11}
	r := GetBufferReader(buf)

	val := r.ReadBits(33)
	assert.Equal(t, val, 474857574, "value should be 474857574")

	val = r.ReadBits(6)
	assert.Equal(t, val, 63, "value should be 63")

	val = r.ReadBits(9)
	assert.Equal(t, val, 17, "value should be 17")
}

func TestBsWriter(t *testing.T) {
	writer := GetBufferWriter(3)

	writer.writeBits(0x47, 8)
	writer.writeBits(0, 1)
	writer.writeBits(0, 1)
	writer.writeBits(0, 1)
	writer.writeBits(33, 13)

	expected := []byte{0x47, 0x00, 0x21}
	assert.Equal(t, writer.GetBuf(), expected, "Expected bytes: [0x47, 0x00, 0x21]")
}
