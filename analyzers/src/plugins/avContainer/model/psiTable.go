package model

import (
	"errors"
	"fmt"
)

func PsiTable(manager PsiManager, pktCnt int, pid int, inBuf []byte) (TableStruct, error) {
	pFieldLen := int(inBuf[0])
	tableId := int(inBuf[pFieldLen+1])
	buf := inBuf[(pFieldLen + 2):]
	switch tableId {
	case 0:
		return PatTable(manager, pktCnt, buf)
	case 2:
		return PmtTable(manager, pktCnt, buf)
	case 0xfc:
		return Scte35Table(manager, pktCnt, pid, buf)
	default:
		return nil, errors.New(fmt.Sprintf("Table with tableId %d is not implemented", tableId))
	}
}
