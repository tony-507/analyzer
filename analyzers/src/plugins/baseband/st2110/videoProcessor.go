package st2110

import (
	"math"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/common/io"
)

// Video handling logic for processor struct

func (p *processor) processVideoPackets() {
	pktBuffer := [][]byte{}
	nextImageIdx := -1
	for idx, pkt := range p.buffer {
		if pkt.marker {
			nextImageIdx = idx + 1
			break
		}
		pktBuffer = append(pktBuffer, pkt.payload)
	}

	outRtp := p.buffer[0].timestamp

	if nextImageIdx != -1 {
		if len(p.buffer) > nextImageIdx {
			p.buffer = p.buffer[nextImageIdx:]
		} else {
			p.buffer = []rtpPacket{}
		}
	} else {
		return
	}

	nextRtp := p.buffer[0].timestamp

	p.processVideo(outRtp, nextRtp, pktBuffer)
}

func (p *processor) processVideo(outRtp uint32, nextRtp uint32, pktBuffer [][]byte) {
	// ST-2110-20
	buf := common.MakeSimpleBuf([]byte{})

	buf.SetField("pid", -1, true)
	buf.SetField("streamType", 235, true)
	buf.SetField("pts", outRtp * 300, false)
	buf.SetField("dts", outRtp * 300, false)

	unit := common.NewMediaUnit(buf, common.VIDEO_UNIT)

	field := false
	for _, pkt := range pktBuffer {
		r := io.GetBufferReader(pkt)
		r.ReadBits(16) // Extended sequence number
		continuation := true
		srdLengths := []int{}

		for continuation {
			// TODO: Understand these terms here
			srdLengths = append(srdLengths, r.ReadBits(16))
			field = r.ReadBits(1) != 0
			r.ReadBits(15) // SRD row number
			continuation = r.ReadBits(1) != 0
			r.ReadBits(15) // TODO: What is SRD offset?
		}

		// Sanity: We should be able to read the line
		totalSrdLength := 0
		for _, srdLength := range srdLengths {
			totalSrdLength += srdLength
		}
		if len(r.GetRemainedBuffer()) != totalSrdLength {
			p.logger.Error(
				"Error in reading sample row data: Remaining size %d but expected size %d",
				len(r.GetRemainedBuffer()), totalSrdLength,
			)
		}
	}

	rtpDiff := nextRtp - outRtp

	p.vInfo.field = p.vInfo.field || field
	p.vInfo.isSegmented = p.vInfo.isSegmented || (rtpDiff == 0)
	p.info.maxRtpDiff = int(math.Max(float64(p.info.maxRtpDiff), float64(rtpDiff)))

	if p.delay == 0 {
		vmd := unit.GetVideoData()

		vmd.FrameRate = common.FrameRate(p.vInfo.fr_num, p.vInfo.fr_den)

		if p.vInfo.field {
			if field {
				vmd.PicFlag = common.BOTTOM_FIELD
			} else {
				vmd.PicFlag = common.TOP_FIELD
			}
		} else {
			vmd.PicFlag = common.FRAME
		}
	} else {
		p.delay--

		if p.delay == 0 {
			packetDuration := common.MatchValueInList(
				p.info.maxRtpDiff, []int{1501, 1800}, 100,
			)
			switch packetDuration {
			case 1501:
				p.vInfo.fr_num = 30000
				p.vInfo.fr_den = 1001
			case 1800:
				p.vInfo.fr_num = 25
				p.vInfo.fr_den = 1
			default:
				p.logger.Error(
					"Fail to probe frame rate with RTP diff %d. Retry after 3 frames.",
					packetDuration,
				)
				p.delay += 3
			}

			if !p.vInfo.field || p.vInfo.isSegmented {
				p.vInfo.fr_num *= 2
			}

			if p.delay == 0 {
				p.logger.Info(
					"Probed video frame rate %d/%d, field: %v, isSegmented: %v",
					p.vInfo.fr_num, p.vInfo.fr_den, p.vInfo.field, p.vInfo.isSegmented,
				)
			}
		}
	}

	p.callback.DeliverData(unit)
}
