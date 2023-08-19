package spectrumClient

/****************************************************************
	MARKER FUNCTIONS
****************************************************************/

const (
	divisor = float32(9.18) // TODO: calculatee a mathematical value
)

var (
	frequencyCentre = map[string]float32{
		"10491.50 / 00": 103 / divisor,
		"10492.75 / 01": 230 / divisor,
		"10493.00 / 02": 256 / divisor,
		"10493.25 / 03": 281 / divisor,
		"10493.50 / 04": 307 / divisor,
		"10493.75 / 05": 332 / divisor,
		"10494.00 / 06": 358 / divisor,
		"10494.25 / 07": 383 / divisor,
		"10494.50 / 08": 409 / divisor,
		"10494.75 / 09": 434 / divisor,
		"10495.00 / 10": 460 / divisor,
		"10495.25 / 11": 485 / divisor,
		"10495.50 / 12": 511 / divisor,
		"10495.75 / 13": 536 / divisor,
		"10496.00 / 14": 562 / divisor,
		"10496.25 / 15": 588 / divisor,
		"10496.50 / 16": 613 / divisor,
		"10496.75 / 17": 639 / divisor,
		"10497.00 / 18": 664 / divisor,
		"10497.25 / 19": 690 / divisor,
		"10497.50 / 20": 715 / divisor,
		"10497.75 / 21": 741 / divisor,
		"10490.00 / 22": 767 / divisor,
		"10498.25 / 23": 792 / divisor,
		"10498.50 / 24": 818 / divisor,
		"10498.75 / 25": 843 / divisor,
		"10499.00 / 26": 869 / divisor,
		"10499.25 / 27": 894 / divisor,
	}

	symbolRateWidth = map[string]float32{ // TODO: calculatee a mathematical values
		"2000": 20,
		"1500": 15,
		"1000": 10,
		"500":  8,
		"333":  5,
		"250":  4,
		"125":  3,
		"66":   2,
		"33":   1.5,
	}
)

// Returns frequency and bandWidth Markers as float32
func getMarkers(frequency, sysmbolRate string) (float32, float32) {
	centre := frequencyCentre[frequency]
	width := symbolRateWidth[sysmbolRate]
	return centre, width
}

// TODO: implement CalibratetionPoints()
/*
func CalibratetionPoints() {
	var yp [918]float32

	for _, v := range CalibrationMarkerWidth {
		// yp[v] = 100
		mylogger.Info.Printf("CalibratetionPoints %v", v)
	}

	for i, v := range yp {
		mylogger.Info.Printf("CalibratetionPoints %v  %v", i, v)
	}

}
*/
