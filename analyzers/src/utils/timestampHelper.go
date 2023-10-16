package utils

import (
	"errors"
	"fmt"
	"time"
)

// All calculations are done with 27MHz clock

var RTP_TIMESTAMP_LOOP_POINT = 4294967295 * Clk90k
var TIMESTAMP_DIFF_THRESHOLD = 60 * Second

type MpegClk int64

const (
	Clk27M MpegClk = 1
	Clk90k         = 300 * Clk27M
	Second         = MpegClk(27000000) * Clk27M
	Minute         = 60 * Second
	Hour           = 60 * Minute
	Day            = 24 * Hour
)

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

func (tc *TimeCode) Equals(other *TimeCode) bool {
	return tc.Hour == other.Hour && tc.Minute == other.Minute && tc.Second == other.Second && tc.Frame == other.Frame
}

func GetNextTimeCode(curTc *TimeCode, fr_num int, fr_den int, dropFrame bool) TimeCode {
	maxFrame := (fr_num + fr_den / 2) / fr_den - 1 // Adjustment may be slightly < 0.5
	rv := *curTc

	rv.Frame++
	if curTc.Frame == maxFrame {
		rv.Second++
		rv.Frame = 0
	}
	if rv.Second == 60 {
		rv.Minute++
		rv.Second = 0
	}
	if rv.Minute == 60 {
		rv.Hour++
		rv.Minute = 0
	}
	if rv.Hour == 24 {
		rv.Hour = 0
	}
	if dropFrame && rv.Frame <= 2 && rv.Second == 0 && rv.Minute % 10 > 0 {
		rv.Frame = 2
	}

	return rv
}

func rtpTimestampToUtcInMs(rtp MpegClk, curUtcSec int64) (MpegClk, error) {
	if curUtcSec == -1 {
		curUtcSec = time.Now().Unix()
	}
	curUtc := MpegClk(curUtcSec) * Second
	curRtp := curUtc % RTP_TIMESTAMP_LOOP_POINT
	nLoop := int64(curUtc / RTP_TIMESTAMP_LOOP_POINT)

	if RTP_TIMESTAMP_LOOP_POINT - rtp < TIMESTAMP_DIFF_THRESHOLD && curRtp < TIMESTAMP_DIFF_THRESHOLD {
		nLoop--
	} else if RTP_TIMESTAMP_LOOP_POINT - curRtp < TIMESTAMP_DIFF_THRESHOLD && rtp < TIMESTAMP_DIFF_THRESHOLD {
		nLoop++
	}
	convertedUtc := MpegClk(nLoop) * RTP_TIMESTAMP_LOOP_POINT + rtp

	if convertedUtc - curUtc > TIMESTAMP_DIFF_THRESHOLD || curUtc - convertedUtc > TIMESTAMP_DIFF_THRESHOLD {
		return convertedUtc, errors.New(fmt.Sprintf("> 1 minute gap between actual UTC (%v ms) and converted UTC (%v ms)", curUtcSec * 1000, int64(convertedUtc / 300 / 90)))
	}
	return convertedUtc, nil
}

func getNDFTimeCode(realTime MpegClk, fr_num int, fr_den int, field bool) TimeCode {
	timeOfDay := realTime % Day
	frameDuration := MpegClk(27000000 / fr_num * fr_den)
	tc := TimeCode{
		DropFrame: false,
		Field: true,
	}

	tc.Hour = int(timeOfDay / Hour)
	tc.Minute = int(timeOfDay / Minute % 60)
	tc.Second = int(timeOfDay / Second % 60)
	tc.Frame = int(timeOfDay % Second / frameDuration)
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

func computeLastSyncTime(realTime MpegClk, fr_num int, fr_den int, dailySyncTime MpegClk) MpegClk {
	timeOfDay := realTime % Day
	nHourInDay := timeOfDay / Hour
	calendarDate := realTime - timeOfDay

	if (nHourInDay < dailySyncTime) {
		calendarDate -= Day
	}

	rv := calendarDate + dailySyncTime
	if (nHourInDay < dailySyncTime) {
		rv -= Day
	}

	return rv
}

func getDFTimeCode(realTime MpegClk, fr_num int, fr_den int, dailySyncTime MpegClk, field bool) TimeCode {
	frameDuration := MpegClk(27000000 / fr_num * fr_den)
	lastSyncTime := computeLastSyncTime(realTime, fr_num, fr_den, dailySyncTime)

	nFramesTilLastSync := lastSyncTime / frameDuration
	nFramesSinceEpoch := realTime / frameDuration
	nFramesSinceLastSync := nFramesSinceEpoch - nFramesTilLastSync

	tc := getDFTimeCodeFromNFrames(int64(nFramesSinceLastSync), fr_num, fr_den, field)

	tc.Hour = (tc.Hour + int(dailySyncTime)) % 24

	return tc
}

func utcTimestampToTimeCode(realTime MpegClk, fr_num int, fr_den int, dailySyncTime MpegClk, field bool) TimeCode {
	if fr_den == 1 {
		return getNDFTimeCode(realTime, fr_num, fr_den, field)
	}
	return getDFTimeCode(realTime, fr_num, fr_den, dailySyncTime, field)
}

func RtpTimestampToTimeCode(rtp MpegClk, curUtcSec int64, fr_num int, fr_den int, field bool, dailySyncTimeInHour int) (TimeCode, error) {
	convertedUtc, err := rtpTimestampToUtcInMs(rtp, curUtcSec)
	return utcTimestampToTimeCode(convertedUtc, fr_num, fr_den, MpegClk(dailySyncTimeInHour) * Hour, field), err
	
}
