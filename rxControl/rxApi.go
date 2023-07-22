package rxControl

import (
	"q100receiver/lmClient"
	"q100receiver/logger"
)

type (
	TuConfig struct {
		Band                 string
		WideFrequency        string
		WideSymbolrate       string
		NarrowFrequency      string
		NarrowSymbolrate     string
		VeryNarrowFrequency  string
		VeryNarrowSymbolRate string
	}
)

var (
	Band       Selector
	SymbolRate Selector
	Frequency  Selector

	IsTuned = false
	IsMuted = false
)

func Intitialize(tuc TuConfig) {
	Band = newSelector(const_BAND_LIST, tuc.Band)
	beaconSymbolRate = newSelector(const_BEACON_SYMBOLRATE_LIST, const_BEACON_SYMBOLRATE_LIST[0])
	beaconFrequency = newSelector(const_BEACON_FREQUENCY_LIST, const_BEACON_FREQUENCY_LIST[0])
	wideSymbolRate = newSelector(const_WIDE_SYMBOLRATE_LIST, tuc.WideSymbolrate)
	wideFrequency = newSelector(const_WIDE_FREQUENCY_LIST, tuc.WideFrequency)
	narrowSymbolRate = newSelector(const_NARROW_SYMBOLRATE_LIST, tuc.NarrowSymbolrate)
	narrowFrequency = newSelector(const_NARROW_FREQUENCY_LIST, tuc.NarrowFrequency)
	veryNarrowSymbolRate = newSelector(const_VERY_NARROW_SYMBOLRATE_LIST, tuc.NarrowSymbolrate)
	veryNarrowFrequency = newSelector(const_VERY_NARROW_FREQUENCY_LIST, tuc.VeryNarrowFrequency)

	switchBand()
}

func Stop() {
	logger.Info("Tuner will stop...")
	if IsTuned {
		lmClient.UnTune()
		IsTuned = false
	}
	logger.Info("Tuner has stopped")
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

func Mute() {
	if IsMuted {
		IsMuted = false
	} else {
		IsMuted = true
	}
}

type Selector struct {
	currIndex int
	lastIndex int
	list      []string
	Value     string
}

func IncBandSelector(st *Selector) {
	if st.currIndex < st.lastIndex {
		st.currIndex++
		st.Value = st.list[st.currIndex]
		switchBand()
	}
}

func DecBandSelector(st *Selector) {
	if st.currIndex > 0 {
		st.currIndex--
		st.Value = st.list[st.currIndex]
		switchBand()
	}
}

func IncSelector(st *Selector) {
	if st.currIndex < st.lastIndex {
		st.currIndex++
		st.Value = st.list[st.currIndex]
		somethingChanged()
	}
}

func DecSelector(st *Selector) {
	if st.currIndex > 0 {
		st.currIndex--
		st.Value = st.list[st.currIndex]
		somethingChanged()
	}
}
