package integration

type testCase struct {
	Source   string       `json:"source"`
	App      []string     `json:"app"`
	Expected expectedProp `json:"expected"`
}

type expectedProp struct {
	Tsa tsaExpectedProp `json:"tsa"`
}

type tsaExpectedProp struct {
	JsonList []int `json:"json"`
	CsvList  []int `json:"csv"`
}