package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tony-507/analyzers/src/plugins/common/clock"
	"github.com/tony-507/analyzers/src/plugins/common/io"
	"github.com/tony-507/analyzers/src/tttKernel"
)

func msToMpegClk(ms int) clock.MpegClk {
	return clock.MpegClk(ms * 90) * clock.Clk90k
}

func TestRtpToUtc(t *testing.T) {
	rtp := 666100238
	curUtc := 1695612755
	utc, err := rtpTimestampToUtcInMs(clock.MpegClk(rtp) * clock.Clk90k, int64(curUtc))
	if err != nil {
		panic(err)
	}
	assert.Equal(t, int64(1695612730), int64(utc / 27000000))
}

func TestRtpToUtcWithRtpPassesLoop(t *testing.T) {
	rtp := 10 * 90000
	curUtc := 12345 * int64(RTP_TIMESTAMP_LOOP_POINT) / 90000 - 10
	utc, err := rtpTimestampToUtcInMs(clock.MpegClk(rtp) * clock.Clk90k, curUtc)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 12345 * int64(RTP_TIMESTAMP_LOOP_POINT) / 90000 + 10 - 37, int64(utc / 27000000))
}

func TestRtpToUtcWithUtcPassesLoop(t *testing.T) {
	rtp := RTP_TIMESTAMP_LOOP_POINT - 10 * clock.Second
	curUtc := 12345 * int64(RTP_TIMESTAMP_LOOP_POINT) / 90000 + 10
	utc, err := rtpTimestampToUtcInMs(rtp, curUtc)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 12345 * int64(RTP_TIMESTAMP_LOOP_POINT) / 90000 - 10 - 37, int64(utc / 27000000))
}

func TestConvert25HzTimeCode(t *testing.T) {
	realTimeInMs := 1695617275880 // 25/9/2023 04:47:55:880
	tc := utcTimestampToTimeCode(msToMpegClk(realTimeInMs), 25, 1, 0, false)
	assert.Equal(t, "04:47:55:22", tc.ToString())
}

func TestConvert50HzTimeCode(t *testing.T) {
	realTimeInMs := 1695617275900 // 25/9/2023 04:47:55:900
	tc := utcTimestampToTimeCode(msToMpegClk(realTimeInMs), 50, 1, 0, true)
	assert.Equal(t, "04:47:55:22", tc.ToString())
}

func TestConvert2997HzTimeCode(t *testing.T) {
	realTimeInMs := 1695617275900 // 25/9/2023 04:47:55:900
	tc := utcTimestampToTimeCode(msToMpegClk(realTimeInMs), 30000, 1001, 0, false)
	assert.Equal(t, "04:47:55:27", tc.ToString())
}

func TestConvert5994HzTimeCode(t *testing.T) {
	realTimeInMs := 1695617275900 // 25/9/2023 04:47:55:900
	tc := utcTimestampToTimeCode(msToMpegClk(realTimeInMs), 60000, 1001, 0, true)
	assert.Equal(t, "04:47:55:27", tc.ToString())
}

func TestGetNextTimeCode(t *testing.T) {
	tc := TimeCode{
		Hour: 23,
		Minute: 59,
		Second: 59,
		Frame: 29,
	}
	expected := TimeCode{
		Hour: 0,
		Minute: 0,
		Second: 0,
		Frame: 0,
	}
	assert.Equal(t, expected, GetNextTimeCode(&tc, 30000, 1001, true))

	tc.Frame = 24
	assert.Equal(t, expected, GetNextTimeCode(&tc, 25, 1, false))
}


func TestMediaUnitWithSimpleBuf(t *testing.T) {
	buf := []byte{1, 2, 3}
	buffer := tttKernel.MakeSimpleBuf(buf)

	buffer.SetField("dummy2", 50, true)

	unit := NewMediaUnit(buffer, UNKNOWN_UNIT)

	// Test if *MediaUnit implements CmUnit
	var _ tttKernel.CmUnit = (*MediaUnit)(unit)

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
	r := io.GetBufferReader([]byte{0x45, 0x4E, 0x47})
	assert.Equal(t, 3, r.GetSize(), "Reader buffer size incorrect")
	assert.Equal(t, "45 4e 47", r.ReadHex(3), "Output not equal")
}

func TestReadChar(t *testing.T) {
	r := io.GetBufferReader([]byte{0x45, 0x4E, 0x47})
	assert.Equal(t, "ENG", r.ReadChar(3), "Output not equal")
}

func TestReadPcr(t *testing.T) {
	buf := []byte{0x0e, 0x26, 0xe0, 0x33, 0x7e, 0x11}
	r := io.GetBufferReader(buf)

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
	r := io.GetBufferReader([]byte{0b00010000, 0b11001010})
	expected := []int{7, 2, 4}
	for _, exp := range expected {
		assert.Equal(t, exp, r.ReadExpGolomb())
	}
}

func TestBsWriter(t *testing.T) {
	writer := io.GetBufferWriter(9)

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
