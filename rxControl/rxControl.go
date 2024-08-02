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

const (
	config_rxBand                 = "Narrow"
	config_rxWideSymbolrate       = "1000"
	config_rxNarrowSymbolrate     = "333"
	config_rxVeryNarrowSymbolRate = "125"
	config_rxWideFrequency        = "10494.75 / 09"
	config_rxNarrowFrequency      = "10499.25 / 27"
	config_rxVeryNarrowFrequency  = "10496.00 / 14"
	config_streamUrl              = "rtmp://rtmp.batc.org.uk/live/" // or some youtube channel
	config_streamKey              = "my-stream-key"
)

type (
	RxData_t struct {
		CurBand        string
		CurSymbolRate  string
		CurFrequency   string
		MarkerCentre   float32
		MarkerWidth    float32
		CurIsTuned     bool
		CurIsStreaming bool
	}
)

var (
	lmCmd              = lmClient.LmCmd_t{}
	lmCmdChan          = make(chan lmClient.LmCmd_t, 1)
	rxData             RxData_t
	rxDataChan         chan<- RxData_t
	bandSelector       selector_t
	symbolRateSelector selector_t
	frequencySelector  selector_t

	isTuned  = false
	isOffset = false
)

func HandleCommands(ctx context.Context, rxCmdChan <-chan RxCmd_t, rxDataCh chan<- RxData_t, lmDataChan chan lmClient.LmData_t) {
	rxDataChan = rxDataCh

	bandSelector = newSelector(const_BAND_LIST, config_rxBand)

	beaconSymbolRate = newSelector(const_BEACON_SYMBOLRATE_LIST, const_BEACON_SYMBOLRATE_LIST[0])
	beaconFrequency = newSelector(const_BEACON_FREQUENCY_LIST, const_BEACON_FREQUENCY_LIST[0])

	wideSymbolRate = newSelector(const_WIDE_SYMBOLRATE_LIST, config_rxWideSymbolrate)
	wideFrequency = newSelector(const_WIDE_FREQUENCY_LIST, config_rxWideFrequency)

	narrowSymbolRate = newSelector(const_NARROW_SYMBOLRATE_LIST, config_rxNarrowSymbolrate)
	narrowFrequency = newSelector(const_NARROW_FREQUENCY_LIST, config_rxNarrowFrequency)

	veryNarrowSymbolRate = newSelector(const_VERY_NARROW_SYMBOLRATE_LIST, config_rxNarrowSymbolrate)
	veryNarrowFrequency = newSelector(const_VERY_NARROW_FREQUENCY_LIST, config_rxVeryNarrowFrequency)

	switchBand()

	go lmClient.ReadLonmyndStatus(ctx, lmCmdChan, lmDataChan)

	for {
		select {
		case <-ctx.Done():
			// if isTuned {
			// 	lmCmd.Type = lmClient.CmdUnTune
			// 	lmCmdChan <- lmCmd
			// 	isTuned = false
			// }
			log.Printf("CANCEL ----- rxControl has cancelled")
			return
		case rxCmd := <-rxCmdChan:
			switch rxCmd {
			case CmdDecBand:
				bandSelector.decBandSelector()
			case CmdIncBand:
				bandSelector.incBandSelector()
			case CmdDecSymbolRate:
				symbolRateSelector.decSelector()
			case CmdIncSymbolRate:
				symbolRateSelector.incSelector()
			case CmdDecFrequency:
				frequencySelector.decSelector()
			case CmdIncFrequency:
				frequencySelector.incSelector()
			case CmdTune:
				setLongmynd()
			case CmdStream:
				toggleStreaming()
			}
			// default:
		}
	}
}

func setLongmynd() {
	if isTuned {
		lmCmd.Type = lmClient.CmdUnTune
		lmCmdChan <- lmCmd
		isTuned = false
	} else {
		lmCmd.Type = lmClient.CmdTune
		lmCmd.FrequencyStr = rxData.CurFrequency
		lmCmd.SymbolRateStr = rxData.CurSymbolRate
		lmCmdChan <- lmCmd
		isTuned = true
	}
	rxData.CurIsTuned = isTuned
	rxDataChan <- rxData
}

func toggleStreaming() {
	if rxData.CurIsStreaming {
		// log.Printf("TODO stop streaming to %v %v", config_streamUrl, config_streamKey)
		rxData.CurIsStreaming = false
	} else {
		// log.Printf("TODO start streaming to %v %v", config_streamUrl, config_streamKey)
		rxData.CurIsStreaming = !true
	}
	rxDataChan <- rxData
}

type selector_t struct {
	currIndex int
	lastIndex int
	list      []string
	value     string
}

func (s *selector_t) incBandSelector() {
	if s.currIndex < s.lastIndex {
		s.currIndex++
		s.value = s.list[s.currIndex]
		switchBand()
	}
}

func (s *selector_t) decBandSelector() {
	if s.currIndex > 0 {
		s.currIndex--
		s.value = s.list[s.currIndex]
		switchBand()
	}
}

func (s *selector_t) incSelector() {
	if s.currIndex < s.lastIndex {
		s.currIndex++
		s.value = s.list[s.currIndex]
		somethingChanged()
	}
}

func (s *selector_t) decSelector() {
	if s.currIndex > 0 {
		s.currIndex--
		s.value = s.list[s.currIndex]
		somethingChanged()
	}
}

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
	CmdDecBand       = 1
	CmdIncBand       = 2
	CmdDecSymbolRate = 3
	CmdIncSymbolRate = 4
	CmdDecFrequency  = 5
	CmdIncFrequency  = 6
	CmdTune          = 7
	CmdStream        = 8
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
	lmCmd.Type = lmClient.CmdUnTune
	lmCmdChan <- lmCmd
	isTuned = false

	rxData.CurBand = bandSelector.value
	rxData.CurSymbolRate = symbolRateSelector.value
	rxData.CurFrequency = frequencySelector.value

	rxData.MarkerCentre = const_frequencyCentre[frequencySelector.value] / 9.18 // NOTE: 9.18 is a temporary kludge
	rxData.MarkerWidth = const_symbolRateWidth[symbolRateSelector.value]
	rxData.CurIsTuned = isTuned
	rxData.CurIsStreaming = isOffset
	rxDataChan <- rxData
}
