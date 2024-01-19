package schema

type TestCase struct {
	Source   string        `json:"source"`
	App      []string      `json:"app"`
	Expected *ExpectedProp `json:"expected"`
}

type ExpectedProp struct {
	Tsa     *TestProp     `json:"tsa"`
	EditCap *TestProp     `json:"editCap"`
}

type TestProp struct {
	File   *[]FileProp
}

type FileProp struct {
	Fname  string `json:"fname"`
	Size   int    `json:"size"`
	Md5sum string `json:"md5sum"`
}

type Md5sumExpectedProp struct {
	Fname  string `json:"fname"`
	Md5sum string `json:"md5sum"`
}
