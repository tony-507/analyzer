package utils

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var RTP_TIMESTAMP_LOOP_POINT uint32 = 4294967295
var TIMESTAMP_DIFF_THRESHOLD uint32 = 60 * 90000
var ONE_MINUTE_IN_MS int = 60 * 1000
var ONE_HOUR_IN_MS int = 60 * ONE_MINUTE_IN_MS
var ONE_DAY_IN_MS int = 24 * ONE_HOUR_IN_MS

type TimeCode struct {
	Hour      int
	Minute    int
	Second    int
	Frame     int
	DropFrame bool
	Field     bool
}

func (tc *TimeCode) ToString() string {
	return fmt.Sprintf("%02d:%02d:%02d:%02d", tc.Hour, tc.Minute, tc.Second, tc.Frame)
}

func rtpTimestampToUtcInMs(rtp uint32, curUtcSec int64) (int64, error) {
	if curUtcSec == -1 {
		curUtcSec = time.Now().Unix()
	}
	curUtc90k := curUtcSec * 90000
	curRtp := uint32(curUtc90k % int64(RTP_TIMESTAMP_LOOP_POINT))
	nLoop := curUtc90k / int64(RTP_TIMESTAMP_LOOP_POINT)

	if RTP_TIMESTAMP_LOOP_POINT - rtp < TIMESTAMP_DIFF_THRESHOLD && curRtp < TIMESTAMP_DIFF_THRESHOLD {
		nLoop--
	} else if RTP_TIMESTAMP_LOOP_POINT - curRtp < TIMESTAMP_DIFF_THRESHOLD && rtp < TIMESTAMP_DIFF_THRESHOLD {
		nLoop++
	}

	convertedUtc90k := nLoop * int64(RTP_TIMESTAMP_LOOP_POINT) + int64(rtp)
	convertedUtcMs := convertedUtc90k * 1000 / 90000
	var err error

	if math.Abs(float64(convertedUtc90k - curUtc90k)) > float64(TIMESTAMP_DIFF_THRESHOLD) {
		err = errors.New(fmt.Sprintf("> 1 minute gap between actual UTC (%v ms) and converted UTC (%v ms)", curUtcSec * 1000, convertedUtcMs))
	}
	return int64(convertedUtcMs), err
}

func getNDFTimeCode(realTimeInMs uint64, fr_num int, fr_den int, field bool) TimeCode {
	timeOfDayInMs := realTimeInMs % uint64(ONE_DAY_IN_MS)
	tc := TimeCode{
		DropFrame: false,
		Field: true,
	}

	tc.Hour = int(timeOfDayInMs) / ONE_HOUR_IN_MS
	tc.Minute = (int(timeOfDayInMs) / ONE_MINUTE_IN_MS) % 60
	tc.Second = (int(timeOfDayInMs) / 1000) % 60
	tc.Frame = (int(timeOfDayInMs) % 1000) * fr_num / fr_den / 1000
	if field {
		if tc.Frame % 2 == 0 {
			tc.Field = false
		}
		tc.Frame /= 2
	}

	return tc
}

func getNumFramesForDFIn1Min(fr_num int, fr_den int) int {
	nFramesIn1Sec := int(fr_num / fr_den) + 1
	// Assume nFramesIn1Sec is even
	return nFramesIn1Sec * 60 - 2
}

func getNumFramesForDFIn10Min(fr_num int, fr_den int, field bool) int {
	return getNumFramesForDFIn1Min(fr_num, fr_den) * 10 + 2
}

func getDFTimeCodeFromNFrames(nFrames int64, fr_num int, fr_den int, field bool) TimeCode {
	tc := TimeCode{
		DropFrame: true,
		Field: true,
	}
	if field {
		nFrames /= 2
		fr_num /= 2
	}

	nFramesIn10Min := getNumFramesForDFIn10Min(fr_num, fr_den, field)
	nFramesIn1Min := getNumFramesForDFIn1Min(fr_num, fr_den)
	nFramesIn1Sec := int(fr_num / fr_den) + 1

	n10MinBlocks := nFrames / int64(nFramesIn10Min)
	nFramesRemained := nFrames - n10MinBlocks * int64(nFramesIn10Min)
	
	// Drop frame, i.e. MM increments by 1 and FF jumps to 2, occurs after every
	// 59 * max FF + (max FF - 1)
	// = 60 * max FF - 1
	// = nFramesIn1Min + 1 frames
	n1MinBlock := (nFramesRemained - 2) / int64(nFramesIn1Min)
	nFramesRemained -= n1MinBlock * int64(nFramesIn1Min)

	tc.Hour = (int(n10MinBlocks) / 6) % 24
	tc.Minute = int(n10MinBlocks % 6) * 10 + int(n1MinBlock)
	tc.Second = int(nFramesRemained / int64(nFramesIn1Sec))
	tc.Frame = int(nFramesRemained % int64(nFramesIn1Sec))

	if tc.Frame <= 2 && tc.Second == 0 && tc.Minute % 10 > 0 {
		tc.Frame = 2
	}

	return tc
}

func computeLastSyncTimeInMs(realTimeInMs uint64, fr_num int, fr_den int, dailySyncTime int) uint64 {
	var lastSyncTime uint64 = 0

	timeOfDayInMs := realTimeInMs % uint64(ONE_DAY_IN_MS)
	nHourInDay := timeOfDayInMs / uint64(ONE_HOUR_IN_MS)
	calendarDateInMs := realTimeInMs - timeOfDayInMs

	if (nHourInDay < uint64(dailySyncTime)) {
		calendarDateInMs -= uint64(ONE_DAY_IN_MS)
	}

	dailySyncTimeInMs := dailySyncTime * ONE_HOUR_IN_MS
	lastSyncTime = calendarDateInMs + uint64(dailySyncTimeInMs)
	if (nHourInDay < uint64(dailySyncTime)) {
		lastSyncTime -= uint64(ONE_DAY_IN_MS)
	}

	// Adjustment when close to next sync time
	nextSyncTime := lastSyncTime + uint64(ONE_DAY_IN_MS)
	if nextSyncTime - realTimeInMs < 100 {
		nFrames := nextSyncTime * uint64(fr_num) / uint64(fr_den) / 1000
		convertedTimeInMs := nFrames * uint64(fr_den) / uint64(fr_num)

		if realTimeInMs >= convertedTimeInMs {
			lastSyncTime = nextSyncTime
		}
	}

	return lastSyncTime
}

func getDFTimeCode(realTimeInMs uint64, fr_num int, fr_den int, dailySyncTime int, field bool) TimeCode {
	lastSyncTime := computeLastSyncTimeInMs(realTimeInMs, fr_num, fr_den, dailySyncTime)

	nFramesTilLastSync := lastSyncTime * uint64(fr_num) / uint64(fr_den) / 1000
	nFramesSinceEpoch := realTimeInMs * uint64(fr_num) / uint64(fr_den) / 1000
	nFramesSinceLastSync := nFramesSinceEpoch - nFramesTilLastSync

	tc := getDFTimeCodeFromNFrames(int64(nFramesSinceLastSync), fr_num, fr_den, field)

	tc.Hour = (tc.Hour + dailySyncTime) % 24

	return tc
}

func utcTimestampToTimeCode(realTimeInMs uint64, fr_num int, fr_den int, dailySyncTime int, field bool) TimeCode {
	if fr_den == 1 {
		return getNDFTimeCode(realTimeInMs, fr_num, fr_den, field)
	}
	return getDFTimeCode(realTimeInMs, fr_num, fr_den, dailySyncTime, field)
}

func RtpTimestampToTimeCode(rtp uint32, curUtcSec int64, fr_num int, fr_den int, field bool, dailySyncTime int) (TimeCode, error) {
	convertedUtcMs, err := rtpTimestampToUtcInMs(rtp, curUtcSec)
	return utcTimestampToTimeCode(uint64(convertedUtcMs), fr_num, fr_den, dailySyncTime, field), err
	
}
