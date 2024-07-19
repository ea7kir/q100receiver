package lmClient

import (
	"fmt"
	"log"
	"strconv"
)

/***********************************************************
	internal to this file
***********************************************************/

const (
	kDash         = "-"
	kInitialising = "Initialising"
	kSeaching     = "Seaching"
	kFoundHeaders = "Found Headers"
	kLocked       = "Locked"
	kDVB_S        = "DVB-S"
	kDVB_S2       = "DVB-S2"
)

type (
	esPairStuct struct {
		waitingFor1stPid  bool
		waitingFor2ndPid  bool
		waitingFor1stType bool
		waitingFor2ndType bool
		the1stPidValue    string
		the1stTypeValue   string
		the2ndPidValue    string
		the2ndTypeValue   string
	}

	agcPairStuct struct {
		waitingForAgc2 bool
		the1stAgcValue int
		the2ndAgcValue int
	}
)

var (
	esPair  = esPairStuct{}
	agcPair = agcPairStuct{}
)

/***********************************************************
	functions called from lmClient.go
***********************************************************/

func (d *esPairStuct) reset() {
	d.waitingFor1stPid = true
	d.waitingFor2ndPid = false
	d.waitingFor1stType = false
	d.waitingFor2ndType = false
	d.the1stPidValue = kDash
	d.the1stTypeValue = kDash
	d.the2ndPidValue = kDash
	d.the2ndTypeValue = kDash
}

func (d *agcPairStuct) reset() {
	d.waitingForAgc2 = false
	d.the1stAgcValue = 0
	d.the2ndAgcValue = 0
}

func (d *LmData_t) reset() {
	d.resetPartial()
	d.changed = false
	d.State = kDash
}

func (d *LmData_t) resetPartial() {
	d.StatusMsg = "Not tuned"
	// d.State = kDash
	d.Frequency = kDash
	d.SymbolRate = kDash
	d.DbMer = kDash
	d.Provider = kDash
	d.Service = kDash
	d.NullRatio = kDash
	d.PidPair1 = kDash
	d.PidPair2 = kDash
	d.VideoCodec = kDash
	d.AudioCodec = kDash
	d.Constellation = kDash
	d.Fec = kDash
	d.Mode = kDash
	d.DbMargin = kDash
	d.DbmPower = kDash
	d.FreqOffset = kDash
	d.changed = true
}

// State
func (d *LmData_t) id1_setState(stateStr string) {
	switch stateStr {
	case "0":
		d.State = kInitialising
		// d.isLocked = false
	case "1":
		d.State = kSeaching
		// d.isLocked = false
	case "2":
		d.State = kFoundHeaders
		// d.isLocked = false
	case "3":
		d.State = kLocked
		d.Mode = kDVB_S
		// d.isLocked = true
	case "4":
		d.State = kLocked
		d.Mode = kDVB_S2
		// d.isLocked = true
	default:
		log.Printf("WARN Undefined status: %v", stateStr)
		return
	}
	d.changed = true
}

// Carrier Frequency - During a search this is the carrier frequency being trialled. When locked this is the Carrier Frequency detected in the stream. Sent in KHz
func (d *LmData_t) id6_setFrequency(carrierFrequencyStr string) {
	kHzFloat, err := strconv.ParseFloat(carrierFrequencyStr, 64)
	if err != nil {
		log.Printf("WARN Bad carrierFrequencyStr: %v", err)
		d.Frequency = kDash
		return
	}
	receivedFrequencyKHz := kHzFloat + config_LmOffset
	d.Frequency = fmt.Sprintf("%.2f", receivedFrequencyKHz/1000)

	frequencyErroorKHz := (receivedFrequencyKHz - frequencyRequestedKHz)
	d.FreqOffset = fmt.Sprintf("%.3f", frequencyErroorKHz/1000)
	d.changed = true
}

// Symbol Rate - During a search this is the symbol rate being trialled.  When locked this is the symbol rate detected in the stream
func (d *LmData_t) id9_setSymbolRate(symbolRateStr string) {
	sysmbolRateFloat, err := strconv.ParseFloat(symbolRateStr, 64)
	if err != nil {
		log.Printf("WARN Bad symbolRateStr: %v", err)
		d.SymbolRate = kDash
		return
	}
	sysmbolRate := sysmbolRateFloat / 1000.0
	d.SymbolRate = fmt.Sprintf("%.1f", sysmbolRate)
	d.changed = true
}

// MER - Modulation Error Ratio in dB * 10
func (d *LmData_t) id12_setDbMer(merStr string) {
	dbMerFloat, err := strconv.ParseFloat(merStr, 64)
	if err != nil {
		log.Printf("WARN Bad merStr: %v", err)
		d.DbMer = kDash
		return
	}
	dbMer := dbMerFloat / 10.0
	d.DbMer = fmt.Sprintf("%.1f", dbMer)
	d.changed = true
}

// Service Provider - TS Service Provider Name
func (d *LmData_t) id13_setProvider(providerStr string) {
	if providerStr == "" {
		d.Provider = kDash
		return
	}
	d.Provider = providerStr
	d.changed = true
}

// Service Provider Service - TS Service Name
func (d *LmData_t) id14_setService(serviceStr string) {
	if serviceStr == "" {
		d.Service = kDash
		return
	}
	d.Service = serviceStr
	d.changed = true
}

// Null Ratio - Ratio of Nulls in TS as percentage
func (d *LmData_t) id15_setNullRatio(nullRatioStr string) {
	if nullRatioStr == "" {
		log.Printf("WARN Missing nullRatioStr")
		d.NullRatio = kDash
		return
	}
	d.NullRatio = nullRatioStr
	d.changed = true
}

// The PID numbers themselves are fairly arbitrary, will vary based on the transmitted signal and don't really mean anything in a single program multiplex.
func (d *LmData_t) id16_setEsPid(esPidStr string) {
	// In the status stream 16 and 17 always come in pairs, 16 is the PID and 17 is the type for that PID, e.g.
	// This means that PID 257 is of type 27 which you look up in the table to be H.264 and PID 258 is type 3 which the table says is MP3.
	// $16,257 == PID 257 is of type 27 which you look up in the table to be H.264
	// $17,27  meaning H.264
	// $16,258 == PID 258 is type 3 which the table says is MP3
	// $17,3   meaaning MP3
	// The PID numbers themselves are fairly arbitrary, will vary based on the transmitted signal and don't really mean anything in a single program multiplex.

	if esPair.waitingFor1stPid {
		esPair.the1stPidValue = esPidStr
		esPair.waitingFor1stPid = false
		esPair.waitingFor2ndPid = false
		esPair.waitingFor1stType = true
		esPair.waitingFor2ndType = false
	}
	if esPair.waitingFor2ndPid {
		esPair.the2ndPidValue = esPidStr
		esPair.waitingFor1stPid = false
		esPair.waitingFor2ndPid = false
		esPair.waitingFor1stType = false
		esPair.waitingFor2ndType = true
	}
}

// ES TYPE - Elementary Stream Type (repeated as pair with 16 for each ES)
func (d *LmData_t) id17_setEsType(esType string) {
	typ, err := strconv.Atoi(esType)
	if err != nil {
		log.Printf("WARN Failed to convert esType %v", err)
		return
	}
	if esPair.waitingFor1stType {
		esPair.the1stTypeValue = esType
		esPair.waitingFor1stPid = false
		esPair.waitingFor2ndPid = true
		esPair.waitingFor1stType = false
		esPair.waitingFor2ndType = false
		d.PidPair1 = fmt.Sprintf("%v %v", esPair.the1stPidValue, esPair.the1stTypeValue) // beacon 257 27 = video
		switch typ {
		case 1:
			d.VideoCodec = "MPEG1"
		case 16:
			d.VideoCodec = "H.263"
		case 27:
			d.VideoCodec = "H.264"
		case 33:
			d.VideoCodec = "JPG2K"
		case 36:
			d.VideoCodec = "H.265"
		case 51:
			d.VideoCodec = "H.266"
		default:
			d.VideoCodec = "???"
		}
	}
	if esPair.waitingFor2ndType {
		esPair.the2ndTypeValue = esType
		esPair.waitingFor1stPid = true
		esPair.waitingFor2ndPid = false
		esPair.waitingFor1stType = false
		esPair.waitingFor2ndType = false
		d.PidPair2 = fmt.Sprintf("%v %v", esPair.the2ndPidValue, esPair.the2ndTypeValue) // beacon 258 3 = audio
		switch typ {
		case 2:
			d.AudioCodec = "MPEG2"
		case 3:
			d.AudioCodec = "MPA" // was "MP3"
		case 4:
			d.AudioCodec = "MP3"
		case 15:
			d.AudioCodec = "ACC"
		case 32:
			d.AudioCodec = "MPA"
		case 129:
			d.AudioCodec = "AC3"
		default:
			d.AudioCodec = "???"
		}
		// TODO: should we reset esPair here ?
	}

	d.changed = true
}

// MODCOD - Received Modulation & Coding Rate. See MODCOD Lookup Table below
func (d *LmData_t) id18_setConstellationAndFecAndMargin(modcodStr string) {
	// set Constellation and Fec
	modcodInt, err := strconv.Atoi(modcodStr) // wiil panic panic if modcodInt is > 28
	if err != nil {
		log.Printf("WARN Failed to convert modcodStr %v", err)
		return
	}
	// d.Constellation = kDash
	// d.Fec = kDash
	switch d.Mode {
	case kDVB_S:
		if modcodInt > len(kModcodeDvdS)-1 {
			log.Printf("WARN DVB-S modcodInt (%v) > (%v)", modcodInt, len(kModcodeDvdS)-1) // to avoid panic
			d.Constellation = kDash
			return
		}
		d.Constellation = kModcodeDvdS[modcodInt].constellation
		d.Fec = kModcodeDvdS[modcodInt].fec
	case kDVB_S2:
		if modcodInt > len(kModcodeDvdS2)-1 {
			log.Printf("WARN DVB-S2 modcodInt (%v) > (%v)", modcodInt, len(kModcodeDvdS2)-1) // to avoid panic
			d.Constellation = kDash
			return
		}
		d.Constellation = kModcodeDvdS2[modcodInt].constellation // TODO: throws panic: runtime error: index out of range [31] with length 29
		d.Fec = kModcodeDvdS2[modcodInt].fec
	default:
		log.Printf("WARN Unknkown longmyndData.mode %v", d.Mode) // TODO: why here, when no signal received ?
		return
	}
	// set Margin
	if d.DbMer == kDash || d.Fec == kDash || d.Constellation == kDash {
		log.Printf("WARN Failed to set Margin at this time")
		d.DbMargin = kDash
		return
	}
	//key := "KEY"
	var key string
	switch d.Mode {
	case kDVB_S:
		key = kDVB_S + " " + d.Fec
	case kDVB_S2:
		// TODO: something better than this
		if d.Constellation == "DummyPL" {
			d.DbMargin = "x"
			return
		}
		key = kDVB_S2 + " " + d.Constellation + " " + d.Fec
	default:
		log.Printf("WARN Unknown d.Mode: %v", d.Mode)
		return
	}

	float_threshold, ok := const_ModeFecThreshold[key]
	if !ok {
		log.Printf("WARN const_ModeFecThreshold key not found")
		d.DbMargin = kDash
		return
	}
	float_mer, err := strconv.ParseFloat(d.DbMer, 64)
	if err != nil {
		log.Printf("WARN Bad longmyndData.dbMer: %v", err)
		d.DbMargin = kDash
		return
	}
	d.DbMargin = fmt.Sprintf("D %.1f", float_mer-float_threshold)
	d.changed = true
}

// AGC1 Gain - Gain value of AGC1 (0: Signal too weak, 65535: Signal too strong)
func (d *LmData_t) id26_setDbmPower(agc1Str string) {
	if agcPair.waitingForAgc2 {
		return
	}
	agc1, err := strconv.Atoi(agc1Str)
	if err != nil {
		log.Printf("WARN Failed to convert agc1Str %v", err)
		return
	}
	agcPair.the1stAgcValue = agc1
	agcPair.waitingForAgc2 = true
}

// AGC2 Gain - Gain value of AGC2 (0: Minimum Gain, 65535: Maximum Gain)
func (d *LmData_t) id27_setDbmPower(agc2Str string) {
	if !agcPair.waitingForAgc2 {
		return
	}
	agc2, err := strconv.Atoi(agc2Str)
	if err != nil {
		log.Printf("WARN Failed to convert agc2Str %v", err)
		d.DbmPower = kDash
		return
	}
	agcPair.the2ndAgcValue = agc2

	power := 0
	v := agcPair.the1stAgcValue
	if v > 0 {
		for _, n := range const_Agc1 {
			if n[0] >= v {
				power = n[1]
				break
			}
		}
	} else {
		v = agcPair.the2ndAgcValue
		for _, n := range const_Agc2 {
			if n[0] >= v {
				power = n[1]
				break
			}
		}

	}
	// log.Printf("INFO ----------------------- agc1 %v agc2 %v", agcPair.the1stAgcValue, agcPair.the2ndAgcValue)

	agcPair.reset()
	d.DbmPower = fmt.Sprint(power)
	d.changed = true
}
