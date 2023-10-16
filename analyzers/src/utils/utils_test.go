package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func msToMpegClk(ms int) MpegClk {
	return MpegClk(ms * 90) * Clk90k
}

func TestRtpToUtc(t *testing.T) {
	rtp := 666100238
	curUtc := 1695612755
	utc, err := rtpTimestampToUtcInMs(MpegClk(rtp) * Clk90k, int64(curUtc))
	if err != nil {
		panic(err)
	}
	assert.Equal(t, int64(1695612767), int64(utc / 27000000))
}

func TestRtpToUtcWithRtpPassesLoop(t *testing.T) {
	rtp := 10 * 90000
	curUtc := 12345 * int64(RTP_TIMESTAMP_LOOP_POINT) / 90000 - 10
	utc, err := rtpTimestampToUtcInMs(MpegClk(rtp) * Clk90k, curUtc)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 12345 * int64(RTP_TIMESTAMP_LOOP_POINT) / 90000 + 10, int64(utc / 27000000))
}

func TestRtpToUtcWithUtcPassesLoop(t *testing.T) {
	rtp := RTP_TIMESTAMP_LOOP_POINT - 10 * Second
	curUtc := 12345 * int64(RTP_TIMESTAMP_LOOP_POINT) / 90000 + 10
	utc, err := rtpTimestampToUtcInMs(rtp, curUtc)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 12345 * int64(RTP_TIMESTAMP_LOOP_POINT) / 90000 - 10, int64(utc / 27000000))
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
