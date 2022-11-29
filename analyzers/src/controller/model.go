package controller

// Controller interface
type CtrlInterface struct {
	SourceSetting SourceSetting
	DemuxSetting  DemuxSetting
	OutputSetting OutputSetting
}

// Source model
type SourceSetting struct {
	FileInput FileInputSetting
	SkipCnt   int
	MaxInCnt  int
}

type FileInputSetting struct {
	Fname string
}

// Demuxer model
type DemuxSetting struct {
	Mode int // 1: psi, 2: full
}

// Output model
type OutputSetting struct {
	DataOutput DataOutputSetting
}

type DataOutputSetting struct {
	OutDir string
}
