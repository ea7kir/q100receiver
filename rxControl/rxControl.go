/*
 *  Q-100 Receiver
 *  Copyright (c) 2023 Michael Naylor EA7KIR (https://michaelnaylor.es)
 */

package rxControl

import (
	"log"
	"q100receiver/lmClient"
)

// BEGIN API ****************************************************

type (
	TuConfig_t struct {
		Band                 string
		WideFrequency        string
		WideSymbolrate       string
		NarrowFrequency      string
		NarrowSymbolrate     string
		VeryNarrowFrequency  string
		VeryNarrowSymbolRate string
	}
	TuData_t struct {
		MarkerCentre float32
		MarkerWidth  float32
	}
)

var (
	tuData     TuData_t
	dataChan   *chan TuData_t
	Band       Selector_t
	SymbolRate Selector_t
	Frequency  Selector_t

	IsTuned  = false
	IsOffset = false
)

func Start(cfg TuConfig_t, ch chan TuData_t) {
	dataChan = &ch

	Band = newSelector(const_BAND_LIST, cfg.Band)

	beaconSymbolRate = newSelector(const_BEACON_SYMBOLRATE_LIST, const_BEACON_SYMBOLRATE_LIST[0])
	beaconFrequency = newSelector(const_BEACON_FREQUENCY_LIST, const_BEACON_FREQUENCY_LIST[0])

	wideSymbolRate = newSelector(const_WIDE_SYMBOLRATE_LIST, cfg.WideSymbolrate)
	wideFrequency = newSelector(const_WIDE_FREQUENCY_LIST, cfg.WideFrequency)

	narrowSymbolRate = newSelector(const_NARROW_SYMBOLRATE_LIST, cfg.NarrowSymbolrate)
	narrowFrequency = newSelector(const_NARROW_FREQUENCY_LIST, cfg.NarrowFrequency)

	veryNarrowSymbolRate = newSelector(const_VERY_NARROW_SYMBOLRATE_LIST, cfg.NarrowSymbolrate)
	veryNarrowFrequency = newSelector(const_VERY_NARROW_FREQUENCY_LIST, cfg.VeryNarrowFrequency)

	switchBand()
}

func Stop() {
	log.Printf("INFO Tuner will stop... - NOT IMPLEMENTED")
	if IsTuned {
		lmClient.UnTune()
		IsTuned = false
	}
	log.Printf("INFO Tuner has stopped")
}

func Tune() {
	if IsTuned {
		lmClient.UnTune()
		IsTuned = false
	} else {
		lmClient.Tune(Frequency.Value, SymbolRate.Value)
		IsTuned = true
	}
}

func SetOffset() {
	if IsOffset {
		IsOffset = false
	} else {
		IsOffset = true
	}
}

type Selector_t struct {
	currIndex int
	lastIndex int
	list      []string
	Value     string
}

func IncBandSelector(st *Selector_t) {
	if st.currIndex < st.lastIndex {
		st.currIndex++
		st.Value = st.list[st.currIndex]
		switchBand()
	}
}

func DecBandSelector(st *Selector_t) {
	if st.currIndex > 0 {
		st.currIndex--
		st.Value = st.list[st.currIndex]
		switchBand()
	}
}

func IncSelector(st *Selector_t) {
	if st.currIndex < st.lastIndex {
		st.currIndex++
		st.Value = st.list[st.currIndex]
		somethingChanged()
	}
}

func DecSelector(st *Selector_t) {
	if st.currIndex > 0 {
		st.currIndex--
		st.Value = st.list[st.currIndex]
		somethingChanged()
	}
}

// END API ****************************************************

var (
	const_BAND_LIST = []string{
		"Beacon",
		"Wide",
		"Narrow",
		"V.Narrow",
	}
	const_BEACON_SYMBOLRATE_LIST = []string{
		"1500",
	}
	const_WIDE_SYMBOLRATE_LIST = []string{
		"1000",
		"1500",
		"2000",
	}
	const_NARROW_SYMBOLRATE_LIST = []string{
		"250",
		"333",
		"500",
	}
	const_VERY_NARROW_SYMBOLRATE_LIST = []string{
		"33",
		"66",
		"125",
	}
	const_BEACON_FREQUENCY_LIST = []string{
		"10491.50 / 00",
	}
	const_WIDE_FREQUENCY_LIST = []string{
		"10493.25 / 03",
		"10494.75 / 09",
		"10496.25 / 15",
	}
	const_NARROW_FREQUENCY_LIST = []string{
		"10492.75 / 01",
		"10493.25 / 03",
		"10493.75 / 05",
		"10494.25 / 07",
		"10494.75 / 09",
		"10495.25 / 11",
		"10495.75 / 13",
		"10496.25 / 15",
		"10496.75 / 17",
		"10497.25 / 19",
		"10497.75 / 21",
		"10498.25 / 23",
		"10498.75 / 25",
		"10499.25 / 27", // index 13
	}
	const_VERY_NARROW_FREQUENCY_LIST = []string{
		"10492.75 / 01",
		"10493.00 / 02",
		"10493.25 / 03",
		"10493.50 / 04",
		"10493.75 / 05",
		"10494.00 / 06",
		"10494.25 / 07",
		"10494.50 / 08",
		"10494.75 / 09",
		"10495.00 / 10",
		"10495.25 / 11",
		"10495.50 / 12",
		"10495.75 / 13",
		"10496.00 / 14", // index 13
		"10496.25 / 15",
		"10496.50 / 16",
		"10496.75 / 17",
		"10497.00 / 18",
		"10497.25 / 19",
		"10497.50 / 20",
		"10497.75 / 21",
		"10498.00 / 22",
		"10498.25 / 23",
		"10498.50 / 24",
		"10498.75 / 25",
		"10499.00 / 26",
		"10499.25 / 27",
	}

	beaconSymbolRate     Selector_t
	beaconFrequency      Selector_t
	wideSymbolRate       Selector_t
	narrowSymbolRate     Selector_t
	veryNarrowSymbolRate Selector_t
	wideFrequency        Selector_t
	narrowFrequency      Selector_t
	veryNarrowFrequency  Selector_t
)

func indexInList(list []string, with string) int { // TODO: add error check
	for i := range list {
		if list[i] == with {
			return i
		}
	}
	return 0
}

func newSelector(values []string, with string) Selector_t {
	index := indexInList(values, with)
	st := Selector_t{
		currIndex: index,
		lastIndex: len(values) - 1,
		list:      values,
		Value:     values[index],
	}
	return st
}

func switchBand() { // TODO: should switch back to previosly use settings
	switch Band.Value {
	case const_BAND_LIST[0]: // beacon
		SymbolRate = beaconSymbolRate
		Frequency = beaconFrequency
	case const_BAND_LIST[1]: // wide
		SymbolRate = wideSymbolRate
		Frequency = wideFrequency
	case const_BAND_LIST[2]: // narrow
		SymbolRate = narrowSymbolRate
		Frequency = narrowFrequency
	case const_BAND_LIST[3]: // very narrow
		SymbolRate = veryNarrowSymbolRate
		Frequency = veryNarrowFrequency
	}
	somethingChanged()
}

func somethingChanged() {
	lmClient.UnTune()
	IsTuned = false
	tuData.MarkerCentre = const_frequencyCentre[Frequency.Value] / 9.18 // NOTE: 9.18 is a temporary kludge
	tuData.MarkerWidth = const_symbolRateWidth[SymbolRate.Value]
	*dataChan <- tuData
}
