package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRtpToUtc(t *testing.T) {
	rtp := 666100238
	curUtc := 1695612755
	utc, err := rtpTimestampToUtcInMs(uint32(rtp), int64(curUtc))
	if err != nil {
		panic(err)
	}
	assert.Equal(t, int64(1695612767), utc / 1000)
}

func TestRtpToUtcWithRtpPassesLoop(t *testing.T) {
	rtp := 10 * 90000
	curUtc := 12345 * int64(RTP_TIMESTAMP_LOOP_POINT) / 90000 - 10
	utc, err := rtpTimestampToUtcInMs(uint32(rtp), curUtc)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 12345 * int64(RTP_TIMESTAMP_LOOP_POINT) / 90000 + 10, utc / 1000)
}

func TestRtpToUtcWithUtcPassesLoop(t *testing.T) {
	rtp := RTP_TIMESTAMP_LOOP_POINT - 10 * 90000
	curUtc := 12345 * int64(RTP_TIMESTAMP_LOOP_POINT) / 90000 + 10
	utc, err := rtpTimestampToUtcInMs(uint32(rtp), curUtc)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 12345 * int64(RTP_TIMESTAMP_LOOP_POINT) / 90000 - 10, utc / 1000)
}

func TestConvert25HzTimeCode(t *testing.T) {
	realTimeInMs := 1695617275880 // 25/9/2023 04:47:55:880
	tc := utcTimestampToTimeCode(uint64(realTimeInMs), 25, 1, 0, false)
	assert.Equal(t, "04:47:55:22", tc.ToString())
}

func TestConvert50HzTimeCode(t *testing.T) {
	realTimeInMs := 1695617275900 // 25/9/2023 04:47:55:900
	tc := utcTimestampToTimeCode(uint64(realTimeInMs), 50, 1, 0, true)
	assert.Equal(t, "04:47:55:22", tc.ToString())
}

func TestConvert2997HzTimeCode(t *testing.T) {
	realTimeInMs := 1695617275900 // 25/9/2023 04:47:55:900
	tc := utcTimestampToTimeCode(uint64(realTimeInMs), 30000, 1001, 0, false)
	assert.Equal(t, "04:47:55:27", tc.ToString())
}

func TestConvert5994HzTimeCode(t *testing.T) {
	realTimeInMs := 1695617275900 // 25/9/2023 04:47:55:900
	tc := utcTimestampToTimeCode(uint64(realTimeInMs), 60000, 1001, 0, true)
	assert.Equal(t, "04:47:55:27", tc.ToString())
}
