package model

import (
	"encoding/json"
	"errors"

	"github.com/tony-507/analyzers/src/plugins/common/io"
)

type ebpInfo struct {
	AcquisitionTime  uint64
	ExtPartition     int
	GroupingIds      []int
	SAPType          int
}

func parseEbpInfo(r *io.BsReader) ([]byte, error) {
	ebp := ebpInfo{}

	if r.ReadChar(4) != "EBP0" {
		return []byte{}, errors.New("Format identifier is not EBP0")
	}

	r.ReadBits(1) // Fragment
	r.ReadBits(1) // Segment
	bSAP := r.ReadBits(1) != 0
	bGrouping := r.ReadBits(1) != 0
	bTime := r.ReadBits(1) != 0
	r.ReadBits(1) // Conceal
	r.ReadBits(1)
	bExtension := r.ReadBits(1) != 0
	extPartitionFlag := false

	if bExtension {
		extPartitionFlag = r.ReadBits(1) != 0
		r.ReadBits(7)
	}

	if bSAP {
		ebp.SAPType = r.ReadBits(3)
		r.ReadBits(5)
	}

	if bGrouping {
		hasNext := r.ReadBits(1) != 0
		ebp.GroupingIds = []int{r.ReadBits(7)}
		for hasNext {
			hasNext = r.ReadBits(1) != 0
			ebp.GroupingIds = append(ebp.GroupingIds, r.ReadBits(7))
		}
	}

	if bTime {
		ebp.AcquisitionTime = uint64(r.ReadBits(64))
	}

	if extPartitionFlag {
		ebp.ExtPartition = r.ReadBits(8)
	}

	r.ReadBits(len(r.GetRemainedBuffer()) * 8)

	return json.Marshal(ebp)
}

type privateData struct {
	Tag    int
	Length int
	Data   string // JSON string if known, otherwise hex string
}

func parsePrivateData(r *io.BsReader) (privateData, error) {
	pd := privateData{}

	pd.Tag = r.ReadBits(8)
	pd.Length = r.ReadBits(8)

	var err error
	var data []byte
	switch pd.Tag {
	case 0xDF:
		data, err = parseEbpInfo(r)
	default:
		pd.Data = r.ReadHex(pd.Length)
	}

	if len(data) != 0 {
		pd.Data = string(data)
	}

	return pd, err
}
