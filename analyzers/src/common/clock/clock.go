package clock

type MpegClk int64

const (
	Clk27M MpegClk = 1
	Clk90k         = 300 * Clk27M
	Second         = MpegClk(27000000) * Clk27M
	Minute         = 60 * Second
	Hour           = 60 * Minute
	Day            = 24 * Hour
)
