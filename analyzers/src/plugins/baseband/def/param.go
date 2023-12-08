package def

type INPUT_TYPE int

const (
	ST_2110 INPUT_TYPE = 1
)

type BasebandProcessorParam struct {
	InputType INPUT_TYPE
}
