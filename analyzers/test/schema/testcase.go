package schema

type TestCase struct {
	Source   string        `json:"source"`
	App      []string      `json:"app"`
	Expected *ExpectedProp `json:"expected"`
}

type ExpectedProp struct {
	Tsa *TsaExpectedProp `json:"tsa"`
}

type TsaExpectedProp struct {
	JsonList []string `json:"json"`
	CsvList  []string `json:"csv"`
}