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
}

type FileInputSetting struct {
	Fname string
}

// Demuxer model
type DemuxSetting struct {
}

// Output model
type OutputSetting struct {
	DataOutput DataOutputSetting
}

type DataOutputSetting struct {
	OutDir string
}
