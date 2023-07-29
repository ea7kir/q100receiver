/*
 *  Q-100 Receiver
 *  Copyright (c) 2023 Michael Naylor EA7KIR (https://michaelnaylor.es)
 */

package lmClient

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"q100receiver/logger"
	"strconv"
	"strings"
)

// BEGIN API ********************************************************

// Represenst all the Longmynd status data being receved
type LongmyndData struct {
	StatusMsg     string
	State         string
	Frequency     string
	SymbolRate    string
	DbMer         string
	Provider      string
	Service       string
	NullRatio     string
	PidPair1      string
	PidPair2      string
	VideoCodec    string
	AudioCodec    string
	Constellation string
	Fec           string
	Mode          string
	DbMargin      string
	DbmPower      string
}

type (
	LmConfig struct {
		Folder      string
		Binary      string
		Offset      float64
		StatusFifo  string
		StartScript string
		StopScript  string
	}
	FpConfig struct {
		Binary      string
		TsFifo      string
		Volume      string
		StartScript string
		StopScript  string
	}
)

func Intitialize(lmc LmConfig, fpc FpConfig, ch chan LongmyndData) {
	lmcfg = &lmc
	fpcfg = &fpc
	lmChannel = ch
	killAll()
	go readLongmynd(lmcfg.StatusFifo, lmcfg.Offset, lmChannel)
}

func Stop() {
	logger.Info.Printf("LmReader will stop...")
	// TODO: implement a better way to stop longmynd and ffplay
	killAll()
	logger.Info.Printf("LmReader has stopped")
}

func Tune(frequency, sysmbolRate string) {
	logger.Info.Printf("------ WILL TUNE")
	isTuned = startLongmynd(frequency, sysmbolRate) // TODO: pass arguments
	// isTuned = true
}

func UnTune() {
	logger.Info.Printf("------ WILL UNTUNE")
	// isTuned = stopLongmynd()
	killAll()
	isTuned = false
	// stopFfplay()
}

// END API ********************************************************

func (p *LongmyndData) reset() {
	p.resetPartial()
	p.State = kDash
}

func (p *LongmyndData) resetPartial() {
	p.StatusMsg = "Not tuned"
	// p.State = kDash
	p.Frequency = kDash
	p.SymbolRate = kDash
	p.DbMer = kDash
	p.Provider = kDash
	p.Service = kDash
	p.NullRatio = kDash
	p.PidPair1 = kDash
	p.PidPair2 = kDash
	p.VideoCodec = kDash
	p.AudioCodec = kDash
	p.Constellation = kDash
	p.Fec = kDash
	p.Mode = kDash
	p.DbMargin = kDash
	p.DbmPower = kDash
}

var (
	lmChannel                      chan LongmyndData
	withFrequency, withSysmbolRate string
)

type (
	tupleConstellationAndFecStruct struct {
		constellation string
		fec           string
	}
)

var (
	kModcodeDvdS = [...]tupleConstellationAndFecStruct{
		{"QPSK", "1/2"}, {"QPSK", "2/3"}, {"QPSK", "3/4"},
		{"QPSK", "5/6"}, {"QPSK", "6/7"}, {"QPSK", "7/8"},
	}

	kModcodeDvdS2 = [...]tupleConstellationAndFecStruct{
		{"DummyPL", "x"}, {"QPSK", "1/4"}, {"QPSK", "1/3"}, {"QPSK", "2/5"},
		{"QPSK", "1/2"}, {"QPSK", "3/5"}, {"QPSK", "2/3"}, {"QPSK", "3/4"},
		{"QPSK", "4/5"}, {"QPSK", "5/6"}, {"QPSK", "8/9"}, {"QPSK", "9/10"},
		{"8PSK", "3/5"}, {"8PSK", "2/3"}, {"8PSK", "3/4"}, {"8PSK", "5/6"},
		{"8PSK", "8/9"}, {"8PSK", "9/10"},
		{"16APSK", "2/3"}, {"16APSK", "3/4"}, {"16APSK", "4/5"}, {"16APSK", "5/6"},
		{"16APSK", "8/9"}, {"16APSK", "9/10"}, {"32APSK", "3/4"}, {"32APSK", "4/5"},
		{"32APSK", "5/6"}, {"32APSK", "8/9"}, {"32APSK", "9/10"},
	}

	kModeFecThreshold = map[string]float64{
		"DVB-S 1/2":          1.7,
		"DVB-S 2/3":          3.3,
		"DVB-S 3/4":          4.2,
		"DVB-S 5/6":          5.1,
		"DVB-S 6/7":          5.5,
		"DVB-S 7/8":          5.8,
		"DVB-S2 QPSK 1/4":    -2.3,
		"DVB-S2 QPSK 1/3":    -1.2,
		"DVB-S2 QPSK 2/5":    -0.3,
		"DVB-S2 QPSK 1/2":    1.0,
		"DVB-S2 QPSK 3/5":    2.3,
		"DVB-S2 QPSK 2/3":    3.1,
		"DVB-S2 QPSK 3/4":    4.1,
		"DVB-S2 QPSK 4/5":    4.7,
		"DVB-S2 QPSK 5/6":    5.2,
		"DVB-S2 QPSK 8/9":    6.2,
		"DVB-S2 QPSK 9/10":   6.5,
		"DVB-S2 8PSK 3/5":    5.5,
		"DVB-S2 8PSK 2/3":    6.6,
		"DVB-S2 8PSK 3/4":    7.9,
		"DVB-S2 8PSK 5/6":    9.4,
		"DVB-S2 8PSK 8/9":    10.7,
		"DVB-S2 8PSK 9/10":   11.0,
		"DVB-S2 16APSK 2/3":  9.0,
		"DVB-S2 16APSK 3/4":  10.2,
		"DVB-S2 16APSK 4/5":  11.0,
		"DVB-S2 16APSK 5/6":  11.6,
		"DVB-S2 16APSK 8/9":  12.9,
		"DVB-S2 16APSK 9/10": 13.2,
		"DVB-S2 32APSK 3/4":  12.8,
		"DVB-S2 32APSK 4/5":  13.7,
		"DVB-S2 32APSK 5/6":  14.3,
		"DVB-S2 32APSK 8/9":  15.7,
		"DVB-S2 32APSK 9/10": 16.1,
	}

	kAgc1 = [...][2]int{
		{1, -70},
		{10, -69},
		{21800, -68},
		{25100, -67},
		{27100, -66},
		{28100, -65},
		{28900, -64},
		{29600, -63},
		{30100, -62},
		{30550, -61},
		{31000, -60},
		{31350, -59},
		{31700, -58},
		{32050, -57},
		{32400, -56},
		{32700, -55},
		{33000, -54},
		{33300, -53},
		{33600, -52},
		{33900, -51},
		{34200, -50},
		{34500, -49},
		{34750, -48},
		{35000, -47},
		{35250, -46},
		{35500, -45},
		{35750, -44},
		{36000, -43},
		{36200, -42},
		{36400, -41},
		{36600, -40},
		{36800, -39},
		{37000, -38},
		{37200, -37},
		{37400, -36},
		{37600, -35},
		{37700, -35},
	}

	kAgc2 = [...][2]int{
		{182, -71},
		{200, -72},
		{225, -73},
		{225, -73},
		{255, -74},
		{290, -75},
		{325, -76},
		{360, -77},
		{400, -78},
		{450, -79},
		{500, -80},
		{560, -81},
		{625, -82},
		{700, -83},
		{780, -84},
		{880, -85},
		{1000, -86},
		{1140, -87},
		{1300, -88},
		{1480, -89},
		{1660, -90},
		{1840, -91},
		{2020, -92},
		{2200, -93},
		{2380, -94},
		{2560, -95},
		{2740, -96},
		{3200, -97},
	}
)

var (
	lmcfg *LmConfig
	fpcfg *FpConfig
)

type esPairStuct struct {
	waitingForPid   bool
	waitingForType  bool
	the1stPidValue  int
	the1stTypeValue int
	the2ndPidValue  int
	the2ndTypeValue int
}

func (p *esPairStuct) reset() {
	p.waitingForPid = false
	p.waitingForType = false
	p.the1stPidValue = 0
	p.the1stTypeValue = 0
	p.the2ndPidValue = 0
	p.the2ndTypeValue = 0
}

type agcPairStuct struct {
	waitingForAgc2 bool
	the1stAgcValue int
	the2ndAgcValue int
}

func (p *agcPairStuct) reset() {
	p.waitingForAgc2 = false
	p.the1stAgcValue = 0
	p.the2ndAgcValue = 0
}

func idAndValFromString(s string) (int, string, error) {
	if !strings.HasPrefix(s, "$") || !strings.Contains(s, ",") || !strings.HasSuffix(s, "\n") || len([]rune(s)) < 3 {
		return 0, "", errors.New("invalid line")
	}
	s = strings.TrimPrefix(s, "$")
	s = strings.TrimSuffix(s, "\n")
	a := strings.Split(s, ",")
	i, err := strconv.Atoi(a[0])
	if err != nil {
		return 0, "", errors.New("invalid id")
	}
	return i, a[1], nil
}

/******************************************************
	Receiving and traslating the raw longmynd stream
******************************************************/

const (
	kDash         = "-"
	kInitialising = "Initialising"
	kSeaching     = "Seaching"
	kFoundHeaders = "Found Headers"
	kLocked       = "Locked"
	kDVB_S        = "DVB-S"
	kDVB_S2       = "DVB-S2"
)

var (
	esPair    = new(esPairStuct)
	agcPair   = new(agcPairStuct)
	liveData  = new(LongmyndData)
	cacheData = new(LongmyndData)

	lmPid     int
	ffPlayPid int
	isTuned   bool
)

// readLongmynd reads the longmynd status fifo and translates to formated strings.
// The results are sent to a channel of type LongmyndData. When no valid signal is being
// received, the LongmyndData fileds will be filled with default values - normall a dash.
func readLongmynd(fifoPath string, offset float64, lonymyndChannel chan LongmyndData) {
	liveData.reset()
	cacheData.reset()
	esPair.reset()

	isLocked := false
	isPlaying := false

	lonymyndChannel <- *liveData

	file, err := os.OpenFile(fifoPath, os.O_CREATE, os.ModeNamedPipe)
	if err != nil {
		logger.Warn.Printf("Failed to open '%v' fifo %v: ", fifoPath, err)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	logger.Info.Printf("Decode forever loop has started")

	for {

		rawStr, err := reader.ReadString(10) // delimited by char(10) == LF
		if err != nil {
			//logger.Error("reading fifo: %v", err)
			liveData.reset()
			cacheData.reset()
			lonymyndChannel <- *liveData
			// time.Sleep(100 * time.Millisecond)
			continue
		}

		lmId, lmVal, err := idAndValFromString(rawStr)
		if err != nil {
			logger.Warn.Printf("Returned from idAndValFromString: %v", err)
			continue
		}

		switch lmId {
		case 1: // State
			id1_setState(lmVal)
			isLocked = liveData.State == kLocked
			if !isLocked { // if not locked, reset most status
				liveData.resetPartial()
				cacheData.reset()
				esPair.reset()
				agcPair.reset()
				lonymyndChannel <- *liveData
				// time.Sleep(5 * time.Millisecond)
				continue
			}
		// case 2: // LNA Gain - On devices that have LNA Amplifiers this represents the two gain sent as N, where n = (lna_gain<<5) | lna_vgo. Though not actually linear, n can be usefully treated as a single byte representing the gain of the amplifier
		// case 3: // Puncture Rate - During a search this is the pucture rate that is being trialled. When locked this is the pucture rate detected in the stream. Sent as a single value, n, where the pucture rate is n/(n+1)
		// case 4: // I Symbol Power - Measure of the current power being seen in the I symbols
		// case 5: // Q Symbol Power - Measure of the current power being seen in the Q symbols
		case 6: // Carrier Frequency - During a search this is the carrier frequency being trialled. When locked this is the Carrier Frequency detected in the stream. Sent in KHz
			id6_setFrequency(lmVal, offset)
		// case 7: // I Constellation - Single signed byte representing the voltage of a sampled I point
		// case 8: // Q Constellation - Single signed byte representing the voltage of a sampled Q point
		case 9: // Symbol Rate - During a search this is the symbol rate being trialled.  When locked this is the symbol rate detected in the stream
			id9_setSymbolRate(lmVal)
		// case 10: // Viterbi Error Rate - Viterbi correction rate as a percentage * 100
		// case 11: // BER - Bit Error Rate as a Percentage * 100
		case 12: // MER - Modulation Error Ratio in dB * 10
			id12_setDbMer(lmVal)
		case 13: // Service Provider - TS Service Provider Name
			id13_setProvider(lmVal)
		case 14: // Service Provider Service - TS Service Name
			id14_setService(lmVal)
		case 15: // Null Ratio - Ratio of Nulls in TS as percentage
			id15_setNullRatio(lmVal)
		case 16: // The PID numbers themselves are fairly arbitrary, will vary based on the transmitted signal and don't really mean anything in a single program multiplex.
			id16_setEsPid(lmVal)
		case 17: // ES TYPE - Elementary Stream Type (repeated as pair with 16 for each ES)
			id17_setEsType(lmVal)
		case 18: // MODCOD - Received Modulation & Coding Rate. See MODCOD Lookup Table below
			id18_setConstellationAndFecAndMargin(lmVal)
		// case 19: // Short Frames - 1 if received signal is using Short Frames, 0 otherwise (DVB-S2 only)
		// case 20: // Pilot Symbols - 1 if received signal is using Pilot Symbols, 0 otherwise (DVB-S2 only)
		// case 21: // LDPC Error Count - LDPC Corrected Errors in last frame (DVB-S2 only)
		// case 22: // BCH Error Count - BCH Corrected Errors in last frame (DVB-S2 only)
		// case 23: // BCH Uncorrected - 1 if some BCH-detected errors were not able to be corrected, 0 otherwise (DVB-S2 only)
		// case 24: // LNB Voltage Enabled - 1 if LNB Voltage Supply is enabled, 0 otherwise (LNB Voltage Supply requires add-on board)
		// case 25: // LNB H Polarisation - 1 if LNB Voltage Supply is configured for Horizontal Polarisation (18V), 0 otherwise (LNB Voltage Supply requires add-on board)
		case 26: // AGC1 Gain - Gain value of AGC1 (0: Signal too weak, 65535: Signal too strong)
			id26_setDbmPower(lmVal)
		case 27: // AGC2 Gain - Gain value of AGC2 (0: Minimum Gain, 65535: Maximum Gain)
			id27_setDbmPower(lmVal)
		} // switch

		if isTuned && isLocked && !isPlaying {
			isPlaying = startFfplay()
		}
		if isTuned && !isLocked && isPlaying {
			isPlaying = stopFfplay()
		}
		if !isTuned && isPlaying {
			isLocked = false
			isPlaying = stopFfplay()
			isTuned = stopLongmynd()
		}

		if isLocked {
			liveData.StatusMsg = fmt.Sprintf("%s : %s : %s", liveData.State, liveData.Provider, liveData.Service)
		} else {
			liveData.StatusMsg = liveData.State
		}

		if *liveData != *cacheData {
			lonymyndChannel <- *liveData
			*cacheData = *liveData
		}
	}
	// logger.Info.Printf("lmreader has stopped")
}

/***********************************************************
	functions called from the main switch statement
***********************************************************/

// State
func id1_setState(stateStr string) {
	switch stateStr {
	case "0":
		liveData.State = kInitialising
	case "1":
		liveData.State = kSeaching
	case "2":
		liveData.State = kFoundHeaders
	case "3":
		liveData.State = kLocked
		liveData.Mode = kDVB_S
	case "4":
		liveData.State = kLocked
		liveData.Mode = kDVB_S2
	default:
		logger.Warn.Printf("Undefined status: %v", stateStr)
		// liveData.reset()
		// return
	}
	// if stateStr < "3" { // if not locked, reset most status
	// 	liveData.resetPartial()
	// 	esPair.reset()
	// 	agcPair.reset()
	// }
}

// Carrier Frequency - During a search this is the carrier frequency being trialled. When locked this is the Carrier Frequency detected in the stream. Sent in KHz
func id6_setFrequency(carrierFrequencyStr string, offset float64) {
	kHzFloat, err := strconv.ParseFloat(carrierFrequencyStr, 64)
	if err != nil {
		logger.Warn.Printf("Bad carrierFrequencyStr: %v", err)
		liveData.Frequency = kDash
		return
	}
	frequency := (kHzFloat + offset) / 1000
	liveData.Frequency = fmt.Sprintf("%.2f", frequency)
}

// Symbol Rate - During a search this is the symbol rate being trialled.  When locked this is the symbol rate detected in the stream
func id9_setSymbolRate(symbolRateStr string) {
	sysmbolRateFloat, err := strconv.ParseFloat(symbolRateStr, 64)
	if err != nil {
		logger.Warn.Printf("Bad symbolRateStr: %v", err)
		liveData.SymbolRate = kDash
		return
	}
	sysmbolRate := sysmbolRateFloat / 1000.0
	liveData.SymbolRate = fmt.Sprintf("%.1f", sysmbolRate)
}

// MER - Modulation Error Ratio in dB * 10
func id12_setDbMer(merStr string) {
	dbMerFloat, err := strconv.ParseFloat(merStr, 64)
	if err != nil {
		logger.Warn.Printf("Bad merStr: %v", err)
		liveData.DbMer = kDash
		return
	}
	dbMer := dbMerFloat / 10.0
	liveData.DbMer = fmt.Sprintf("%.1f", dbMer)
}

// Service Provider - TS Service Provider Name
func id13_setProvider(providerStr string) {
	if providerStr == "" {
		liveData.Provider = kDash
		return
	}
	liveData.Provider = providerStr
}

// Service Provider Service - TS Service Name
func id14_setService(serviceStr string) {
	if serviceStr == "" {
		liveData.Service = kDash
		return
	}
	liveData.Service = serviceStr
}

// Null Ratio - Ratio of Nulls in TS as percentage
func id15_setNullRatio(nullRatioStr string) {
	if nullRatioStr == "" {
		logger.Warn.Printf("Missing nullRatioStr")
		liveData.NullRatio = kDash
		return
	}
	liveData.NullRatio = nullRatioStr
}

// The PID numbers themselves are fairly arbitrary, will vary based on the transmitted signal and don't really mean anything in a single program multiplex.
func id16_setEsPid(esPidStr string) {
	// In the status stream 16 and 17 always come in pairs, 16 is the PID and 17 is the type for that PID, e.g.
	// This means that PID 257 is of type 27 which you look up in the table to be H.264 and PID 258 is type 3 which the table says is MP3.
	// $16,257 == PID 257 is of type 27 which you look up in the table to be H.264
	// $17,27  meaning H.264
	// $16,258 == PID 258 is type 3 which the table says is MP3
	// $17,3   meaaning MP3
	// The PID numbers themselves are fairly arbitrary, will vary based on the transmitted signal and don't really mean anything in a single program multiplex.

	// ignore and do nothing
	if esPair.waitingForType {
		return
	}
	pid, err := strconv.Atoi(esPidStr)
	if err != nil {
		logger.Warn.Printf("Failed to convert esPidStr %v", err)
		return
	}
	esPair.the1stTypeValue = pid
	esPair.waitingForType = true
}

// ES TYPE - Elementary Stream Type (repeated as pair with 16 for each ES)
func id17_setEsType(esType string) {
	if !esPair.waitingForType {
		return
	}
	typ, err := strconv.Atoi(esType)
	if err != nil {
		logger.Warn.Printf("Failed to convert esType %v", err)
		return
	}
	esPair.the2ndTypeValue = typ

	liveData.PidPair2 = fmt.Sprintf("%v %v", esPair.the1stTypeValue, esPair.the2ndTypeValue)

	// liveData.PidPair2 = fmt.Sprintf("%v %v", esPair.the2ndPidValue, esPair.the2ndTypeValue)

	logger.Info.Printf("----------------------- PID %v Type %v", esPair.the1stTypeValue, esPair.the2ndTypeValue)

	switch typ {
	case 1:
		liveData.VideoCodec = "MPEG1"
	case 16:
		liveData.VideoCodec = "H.263"
	case 27:
		liveData.VideoCodec = "H.264"
	case 33:
		liveData.VideoCodec = "JPG2K"
	case 36:
		liveData.VideoCodec = "H.265"
	case 51:
		liveData.VideoCodec = "H.266"
	default:
		// liveData.VideoCodec = "???"
	}

	switch typ {
	case 2:
		liveData.AudioCodec = "MPEG2"
	case 3:
		liveData.AudioCodec = "MP3"
	case 4:
		liveData.AudioCodec = "MP3"
	case 15:
		liveData.AudioCodec = "ACC"
	case 32:
		liveData.AudioCodec = "MPA"
	case 129:
		liveData.AudioCodec = "AC3"
	default:
		// liveData.AudioCodec = "???"
	}

	esPair.reset()
}

// MODCOD - Received Modulation & Coding Rate. See MODCOD Lookup Table below
func id18_setConstellationAndFecAndMargin(modcodStr string) {
	// set Constellation and Fec
	modcodInt, err := strconv.Atoi(modcodStr) // wiil panic panic if modcodInt is > 28
	if err != nil {
		logger.Warn.Printf("Failed to convert modcodStr %v", err)
		return
	}
	// liveData.Constellation = kDash
	// liveData.Fec = kDash
	switch liveData.Mode {
	case kDVB_S:
		if modcodInt > len(kModcodeDvdS)-1 {
			logger.Warn.Printf("DVB-S modcodInt (%v) > (%v)", modcodInt, len(kModcodeDvdS)-1) // to avoid panic
			liveData.Constellation = kDash
			return
		}
		liveData.Constellation = kModcodeDvdS[modcodInt].constellation
		liveData.Fec = kModcodeDvdS[modcodInt].fec
	case kDVB_S2:
		if modcodInt > len(kModcodeDvdS2)-1 {
			logger.Warn.Printf("DVB-S2 modcodInt (%v) > (%v)", modcodInt, len(kModcodeDvdS2)-1) // to avoid panic
			liveData.Constellation = kDash
			return
		}
		liveData.Constellation = kModcodeDvdS2[modcodInt].constellation // TODO: throws panic: runtime error: index out of range [31] with length 29
		liveData.Fec = kModcodeDvdS2[modcodInt].fec
	default:
		// logger.Warn.Printf("Unknkown longmyndData.mode %v", mode) // TODO: why here, when no signal received ?
		return
	}
	// set Margin
	if liveData.DbMer == kDash || liveData.Fec == kDash || liveData.Constellation == kDash {
		logger.Warn.Printf("Failed to set Margin at this time")
		liveData.DbMargin = kDash
		return
	}
	//key := "KEY"
	var key string
	switch liveData.Mode {
	case kDVB_S:
		key = kDVB_S + " " + liveData.Fec
	case kDVB_S2:
		// TODO: something better than this
		if liveData.Constellation == "DummyPL" {
			liveData.DbMargin = "x"
			return
		}
		key = kDVB_S2 + " " + liveData.Constellation + " " + liveData.Fec
	default:
		logger.Warn.Printf("Unknown liveData.Mode: %v", liveData.Mode)
		return
	}

	// float_threshold := kModeFecThreshold[key]
	float_threshold, ok := kModeFecThreshold[key]
	if !ok {
		logger.Warn.Printf("kModeFecThreshold key not foundr")
		liveData.DbMargin = kDash
		return
	}
	float_mer, err := strconv.ParseFloat(liveData.DbMer, 64)
	if err != nil {
		logger.Warn.Printf("Bad longmyndData.dbMer: %v", err)
		liveData.DbMargin = kDash
		return
	}
	liveData.DbMargin = fmt.Sprintf("D %.1f", float_mer-float_threshold)
}

// AGC1 Gain - Gain value of AGC1 (0: Signal too weak, 65535: Signal too strong)
func id26_setDbmPower(agc1Str string) {
	if agcPair.waitingForAgc2 {
		return
	}
	agc1, err := strconv.Atoi(agc1Str)
	if err != nil {
		logger.Warn.Printf("Failed to convert agc1Str %v", err)
		return
	}
	agcPair.the1stAgcValue = agc1
	agcPair.waitingForAgc2 = true
}

// AGC2 Gain - Gain value of AGC2 (0: Minimum Gain, 65535: Maximum Gain)
func id27_setDbmPower(agc2Str string) {
	if !agcPair.waitingForAgc2 {
		return
	}
	agc2, err := strconv.Atoi(agc2Str)
	if err != nil {
		logger.Warn.Printf("Failed to convert agc2Str %v", err)
		liveData.DbmPower = kDash
		return
	}
	agcPair.the2ndAgcValue = agc2

	p := 0
	v := agcPair.the1stAgcValue
	if v > 0 {
		for _, n := range kAgc1 {
			if n[0] >= v {
				p = n[1]
				break
			}
		}
	} else {
		v = agcPair.the2ndAgcValue
		for _, n := range kAgc2 {
			if n[0] >= v {
				p = n[1]
				break
			}
		}

	}
	// logger.Info.Printf("----------------------- agc1 %v agc2 %v", agcPair.the1stAgcValue, agcPair.the2ndAgcValue)

	liveData.DbmPower = fmt.Sprint(p)
	agcPair.reset()
}
