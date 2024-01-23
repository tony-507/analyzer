package monitor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tony-507/analyzers/src/plugins/common"
	"github.com/tony-507/analyzers/src/plugins/monitor/impl"
	"github.com/tony-507/analyzers/src/tttKernel"
)

func TestMonitorHeading(t *testing.T) {
	m := newMonitor()
	pts := 1234567
	tc := common.TimeCode{Hour : 1, Minute: 2, Second: 3, Frame: 4}

	buf1 := tttKernel.MakeSimpleBuf([]byte{})
	buf1.SetField("pts", pts, false)
	unit1 := common.NewMediaUnit(buf1, common.VIDEO_UNIT)
	vmd1 := unit1.GetVideoData()
	vmd1.Type = common.I_SLICE
	vmd1.Tc = tc

	buf2 := tttKernel.MakeSimpleBuf([]byte{})
	buf2.SetField("pts", pts, false)
	unit2 := common.NewMediaUnit(buf2, common.VIDEO_UNIT)
	vmd2 := unit2.GetVideoData()
	vmd2.Type = common.I_SLICE
	vmd2.Tc = tc

	m.setParameter(&OutputMonitorParam{
		Redundancy: &impl.RedundancyParam{
			TimeRef: impl.Vitc,
		},
	})
	m.start()
	m.feed(unit1, "abc")
	m.feed(unit2, "def")
	m.feed(unit1, "abc") // Unit from existing source should not create new columns
	m.stop()

	assert.Equal(t, fmt.Sprintf("|%15s|%15s|%15s|%15s", "PTS_1", "VITC_1", "PTS_2", "VITC_2"), m.heading)
}

func TestRedundancyMonitorVitcMode(t *testing.T) {
	rm := impl.GetRedundancyMonitor(
		&impl.RedundancyParam{
			TimeRef: impl.Vitc,
		},
	)

	pts1 := 1234567
	tc1 := common.TimeCode{Hour : 1, Minute: 2, Second: 3, Frame: 4}
	pts2 := 2234567
	tc2 := common.TimeCode{Hour : 1, Minute: 2, Second: 3, Frame: 6}

	expected := []string{}

	for i := 0; i < 10; i++ {
		buf1 := tttKernel.MakeSimpleBuf([]byte{})
		buf1.SetField("pts", pts1, false)
		unit1 := common.NewMediaUnit(buf1, common.VIDEO_UNIT)
		vmd1 := unit1.GetVideoData()
		vmd1.Type = common.I_SLICE
		vmd1.Tc = tc1

		buf2 := tttKernel.MakeSimpleBuf([]byte{})
		buf2.SetField("pts", pts2, false)
		unit2 := common.NewMediaUnit(buf2, common.VIDEO_UNIT)
		vmd2 := unit2.GetVideoData()
		vmd2.Type = common.I_SLICE
		vmd2.Tc = tc2

		expected = append([]string{fmt.Sprintf("|%15d|%15s|%15d|%15s|", pts1, tc1.ToString(), pts2, tc2.ToString())}, expected...)

		rm.Feed(unit1, "abc")
		rm.Feed(unit2, "def")

		pts1 += 3003
		tc1 = common.GetNextTimeCode(&tc1, 30000, 1001, true)
		pts2 += 3003
		tc2 = common.GetNextTimeCode(&tc2, 30000, 1001, true)
	}
	assert.Equal(t, expected, rm.GetDisplayData())
}

func TestRedundancyMonitorPtsMode(t *testing.T) {
	rm := impl.GetRedundancyMonitor(
		&impl.RedundancyParam{
			TimeRef: impl.Pts,
		},
	)

	pts1 := 1234567
	pts2 := 2234567

	expected := []string{}

	for i := 0; i < 10; i++ {
		buf1 := tttKernel.MakeSimpleBuf([]byte{})
		buf1.SetField("pts", pts1, false)
		unit1 := common.NewMediaUnit(buf1, common.VIDEO_UNIT)
		vmd1 := unit1.GetVideoData()
		vmd1.Type = common.I_SLICE

		buf2 := tttKernel.MakeSimpleBuf([]byte{})
		buf2.SetField("pts", pts2, false)
		unit2 := common.NewMediaUnit(buf2, common.VIDEO_UNIT)
		vmd2 := unit2.GetVideoData()
		vmd2.Type = common.I_SLICE

		expected = append([]string{fmt.Sprintf("|%15d|%15d|", pts1, pts2)}, expected...)

		rm.Feed(unit1, "abc")
		rm.Feed(unit2, "def")

		pts1 += 3003
		pts2 += 3003
	}
	assert.Equal(t, expected, rm.GetDisplayData())
}
