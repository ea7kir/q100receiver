/*
 *  Q-100 Receiver
 *  Copyright (c) 2023 Michael Naylor EA7KIR (https://michaelnaylor.es)
 */

package rxControl

import (
	"context"
	"log"
	"q100receiver/lmClient"
)

// BEGIN API ****************************************************

type (
	RxConfig_t struct {
		Band                 string
		WideFrequency        string
		WideSymbolrate       string
		NarrowFrequency      string
		NarrowSymbolrate     string
		VeryNarrowFrequency  string
		VeryNarrowSymbolRate string
	}
	RxData_t struct {
		CurBand       string
		CurSymbolRate string
		CurFrequency  string
		//
		// FUTURE: received longmynd data could go here
		//
		MarkerCentre float32
		MarkerWidth  float32
		CurIsTuned   bool
		CurIsOffset  bool
	}
)

var (
	rxData             RxData_t
	dataChan           chan RxData_t
	bandSelector       selector_t
	symbolRateSelector selector_t
	frequencySelector  selector_t

	isTuned  = false
	isOffset = false
)

// func HandleUiCommands(ctx, rxConfig, tuChannel) // , tuCmdChan)
func HandleCommands(ctx context.Context, cfg RxConfig_t, cmdCh chan RxCmd_t, dataCh chan RxData_t) {
	dataChan = dataCh

	bandSelector = newSelector(const_BAND_LIST, cfg.Band)

	beaconSymbolRate = newSelector(const_BEACON_SYMBOLRATE_LIST, const_BEACON_SYMBOLRATE_LIST[0])
	beaconFrequency = newSelector(const_BEACON_FREQUENCY_LIST, const_BEACON_FREQUENCY_LIST[0])

	wideSymbolRate = newSelector(const_WIDE_SYMBOLRATE_LIST, cfg.WideSymbolrate)
	wideFrequency = newSelector(const_WIDE_FREQUENCY_LIST, cfg.WideFrequency)

	narrowSymbolRate = newSelector(const_NARROW_SYMBOLRATE_LIST, cfg.NarrowSymbolrate)
	narrowFrequency = newSelector(const_NARROW_FREQUENCY_LIST, cfg.NarrowFrequency)

	veryNarrowSymbolRate = newSelector(const_VERY_NARROW_SYMBOLRATE_LIST, cfg.NarrowSymbolrate)
	veryNarrowFrequency = newSelector(const_VERY_NARROW_FREQUENCY_LIST, cfg.VeryNarrowFrequency)

	switchBand()

	for {
		select {
		case <-ctx.Done():
			if isTuned {
				lmClient.UnTune()
				isTuned = false
			}
			log.Printf("CANCEL ----- rxControl has cancelled")
			return
		case command := <-cmdCh:
			switch command {
			case CmdDecBand:
				decBandSelector(&bandSelector)
			case CmdIncBand:
				incBandSelector(&bandSelector)
			case CmdDecSymbolRate:
				decSelector(&symbolRateSelector)
			case CmdIncSymbolRate:
				incSelector(&symbolRateSelector)
			case CmdDecFrequency:
				decSelector(&frequencySelector)
			case CmdIncFrequency:
				incSelector(&frequencySelector)
			case CmdTune:
				setLongmynd()
			case CmdCalibrate:
				setOffset()
			}
		}
	}
}

// func Stop() {
// 	log.Printf("INFO Tuner will stop... - NOT IMPLEMENTED")
// 	if isTuned {
// 		lmClient.UnTune()
// 		isTuned = false
// 	}
// 	log.Printf("INFO Tuner has stopped")
// }

func setLongmynd() {
	if isTuned {
		lmClient.UnTune()
		isTuned = false
	} else {
		lmClient.Tune(rxData.CurFrequency, rxData.CurSymbolRate) // Frequency.value, SymbolRate.value)
		isTuned = true
	}
	rxData.CurIsTuned = isTuned
	dataChan <- rxData
}

func setOffset() {
	if isOffset {
		isOffset = false
	} else {
		isOffset = true
	}
	rxData.CurIsOffset = isOffset
	dataChan <- rxData
}

type selector_t struct {
	currIndex int
	lastIndex int
	list      []string
	value     string
}

func incBandSelector(st *selector_t) {
	if st.currIndex < st.lastIndex {
		st.currIndex++
		st.value = st.list[st.currIndex]
		switchBand()
	}
}

func decBandSelector(st *selector_t) {
	if st.currIndex > 0 {
		st.currIndex--
		st.value = st.list[st.currIndex]
		switchBand()
	}
}

func incSelector(st *selector_t) {
	if st.currIndex < st.lastIndex {
		st.currIndex++
		st.value = st.list[st.currIndex]
		somethingChanged()
	}
}

func decSelector(st *selector_t) {
	if st.currIndex > 0 {
		st.currIndex--
		st.value = st.list[st.currIndex]
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

	beaconSymbolRate     selector_t
	beaconFrequency      selector_t
	wideSymbolRate       selector_t
	narrowSymbolRate     selector_t
	veryNarrowSymbolRate selector_t
	wideFrequency        selector_t
	narrowFrequency      selector_t
	veryNarrowFrequency  selector_t
)

type RxCmd_t int

const (
	CmdDecBand = iota
	CmdIncBand
	CmdDecSymbolRate
	CmdIncSymbolRate
	CmdDecFrequency
	CmdIncFrequency
	CmdTune
	CmdCalibrate
)

func indexInList(list []string, with string) int { // TODO: add error check
	for i := range list {
		if list[i] == with {
			return i
		}
	}
	return 0
}

func newSelector(values []string, with string) selector_t {
	index := indexInList(values, with)
	st := selector_t{
		currIndex: index,
		lastIndex: len(values) - 1,
		list:      values,
		value:     values[index],
	}
	return st
}

func switchBand() { // TODO: should switch back to previosly use settings
	switch bandSelector.value { // {Band.value {
	case const_BAND_LIST[0]: // beacon
		symbolRateSelector = beaconSymbolRate
		frequencySelector = beaconFrequency
	case const_BAND_LIST[1]: // wide
		symbolRateSelector = wideSymbolRate
		frequencySelector = wideFrequency
	case const_BAND_LIST[2]: // narrow
		symbolRateSelector = narrowSymbolRate
		frequencySelector = narrowFrequency
	case const_BAND_LIST[3]: // very narrow
		symbolRateSelector = veryNarrowSymbolRate
		frequencySelector = veryNarrowFrequency
	}
	somethingChanged()
}

func somethingChanged() {
	lmClient.UnTune()
	isTuned = false

	rxData.CurBand = bandSelector.value
	rxData.CurSymbolRate = symbolRateSelector.value
	rxData.CurFrequency = frequencySelector.value

	rxData.MarkerCentre = const_frequencyCentre[frequencySelector.value] / 9.18 // NOTE: 9.18 is a temporary kludge
	rxData.MarkerWidth = const_symbolRateWidth[symbolRateSelector.value]
	rxData.CurIsTuned = isTuned
	rxData.CurIsOffset = isOffset
	dataChan <- rxData
}
