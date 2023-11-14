package monitor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/monitor/impl"
	"github.com/tony-507/analyzers/src/utils"
)

func TestMonitorHeading(t *testing.T) {
	m := newMonitor()
	pts := 1234567
	tc := utils.TimeCode{Hour : 1, Minute: 2, Second: 3, Frame: 4}

	buf1 := common.MakeSimpleBuf([]byte{})
	buf1.SetField("pts", pts, false)
	buf1.SetField("timecode", tc.ToString(), false)
	unit1 := common.MakeIOUnit(buf1, -1, -1)

	buf2 := common.MakeSimpleBuf([]byte{})
	buf2.SetField("pts", pts, false)
	buf2.SetField("timecode", tc.ToString(), false)
	unit2 := common.MakeIOUnit(buf2, -1, -1)

	m.start()
	m.feed(unit1, "abc")
	m.feed(unit2, "def")
	m.feed(unit1, "abc") // Unit from existing source should not create new columns
	m.stop()

	assert.Equal(t, fmt.Sprintf("|%15s|%15s|%15s|%15s", "PTS_1", "VITC_1", "PTS_2", "VITC_2"), m.heading)
}

func TestRedundancyMonitorDisplay(t *testing.T) {
	rm := impl.GetRedundancyMonitor()

	pts1 := 1234567
	tc1 := utils.TimeCode{Hour : 1, Minute: 2, Second: 3, Frame: 4}
	pts2 := 2234567
	tc2 := utils.TimeCode{Hour : 1, Minute: 2, Second: 3, Frame: 6}

	expected := []string{}

	for i := 0; i < 10; i++ {
		buf1 := common.MakeSimpleBuf([]byte{})
		buf1.SetField("pts", pts1, false)
		buf1.SetField("timecode", tc1.ToString(), false)
		unit1 := common.MakeIOUnit(buf1, -1, -1)

		buf2 := common.MakeSimpleBuf([]byte{})
		buf2.SetField("pts", pts2, false)
		buf2.SetField("timecode", tc2.ToString(), false)
		unit2 := common.MakeIOUnit(buf2, -1, -1)

		expected = append(expected, fmt.Sprintf("|%15d|%15s|%15d|%15s|", pts1, tc1.ToString(), pts2, tc2.ToString()))

		rm.Feed(unit1, "abc")
		rm.Feed(unit2, "def")

		pts1 += 3003
		tc1 = utils.GetNextTimeCode(&tc1, 30000, 1001, true)
		pts2 += 3003
		tc2 = utils.GetNextTimeCode(&tc2, 30000, 1001, true)
	}
	assert.Equal(t, expected, rm.GetDisplayData())
}
