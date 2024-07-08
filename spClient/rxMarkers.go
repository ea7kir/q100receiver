package spClient

/*****************************************************************
* SPECTRUM MARKERS FOR RECEIVING
*****************************************************************/

var (
	// TODO: calculatee a mathematical values
	const_frequencyCentre = map[string]float32{
		"10491.50 / 00": 103,
		"10492.75 / 01": 230,
		"10493.00 / 02": 256,
		"10493.25 / 03": 281,
		"10493.50 / 04": 307,
		"10493.75 / 05": 332,
		"10494.00 / 06": 358,
		"10494.25 / 07": 383,
		"10494.50 / 08": 409,
		"10494.75 / 09": 434,
		"10495.00 / 10": 460,
		"10495.25 / 11": 485,
		"10495.50 / 12": 511,
		"10495.75 / 13": 536,
		"10496.00 / 14": 562,
		"10496.25 / 15": 588,
		"10496.50 / 16": 613,
		"10496.75 / 17": 639,
		"10497.00 / 18": 664,
		"10497.25 / 19": 690,
		"10497.50 / 20": 715,
		"10497.75 / 21": 741,
		"10490.00 / 22": 767,
		"10498.25 / 23": 792,
		"10498.50 / 24": 818,
		"10498.75 / 25": 843,
		"10499.00 / 26": 869,
		"10499.25 / 27": 894,
	}

	// TODO: calculatee a mathematical values
	const_symbolRateWidth = map[string]float32{
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
func getMarkers(frequency, symbolRate string) (float32, float32) {
	centre := const_frequencyCentre[frequency] / 9.18 // NOTE: 9.18 is a temporary kludge
	width := const_symbolRateWidth[symbolRate]
	return centre, width
}

// Sets the spData Marker values
func SetMarker(frequency string, symbolRate string) {
	spData.MarkerCentre, spData.MarkerWidth = getMarkers(frequency, symbolRate)
}
