/*
 *  Q-100 Receiver
 *  Copyright (c) 2023 Michael Naylor EA7KIR (https://michaelnaylor.es)
 */

package lmClient

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	config_LmBaseFolder = "/home/pi/Q100/"

	config_LmFolder     = config_LmBaseFolder + "longmynd/"
	config_LmBinary     = config_LmBaseFolder + "longmynd/longmynd"
	config_LmStatusFifo = config_LmBaseFolder + "longmynd/longmynd_main_status"
	config_LmOffset     = float64(9750000)

	config_FpTsFifo = config_LmBaseFolder + "longmynd/longmynd_main_ts"
	config_FpBinary = "/usr/bin/ffplay"
	config_FpVolume = "100"

	CmdTune          = 1
	CmdUnTune        = 2
	CmdEnableOffset  = 3
	CmdDisableOffset = 4
)

type (
	// LmCmd_t int

	LmCmd_t struct {
		Type          int
		FrequencyStr  string
		FrequencyKHz  float64
		SymbolRateStr string
	}

	LmData_t struct {
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
		FreqOffset    string
	}
)

var (
	frequencyRequestedKHz float64
)

///////////////////////////////////////////////////////////////////////////////////////////

// Reads the longmynd status fifo and translates to formated strings.
//
//	The results are sent to a channel of type LongmyndData. When no valid signal is being
//	received, the LongmyndData fileds will be filled with default values - normally a dash.
//
// func ReadLonmyndStatus(ctx context.Context, lmc LmConfig_t, fpc FpConfig_t, ch chan LongmyndData) {
func ReadLonmyndStatus(ctx context.Context, lmCmdChan <-chan LmCmd_t, lmDataChan chan<- LmData_t) {

	liveData.reset()
	cacheData.reset()
	esPair.reset()

	isLocked := false

	lmDataChan <- *liveData

	var file *os.File = nil
	var reader *bufio.Reader = nil
	var fileIsOpen = false

	for {
		if !fileIsOpen {
			select {
			case <-ctx.Done():
				stopFfPlayAndLongmynd()
				// file.Close()
				log.Printf("CANCEL ----- lmClient 1 has cancelled")
				return
			case cmd := <-lmCmdChan:
				switch cmd.Type {
				case CmdTune:
					log.Printf("INFO ------ WILL TUNE for the first time")
					startLongmynd(cmd.FrequencyStr, cmd.SymbolRateStr)
					var err error
					file, err = os.OpenFile(config_LmStatusFifo, os.O_RDONLY, os.ModeNamedPipe)
					if err != nil {
						log.Fatalf("FATAL Failed to open '%v' fifo %v: ", config_LmStatusFifo, err)
					}
					reader = bufio.NewReader(file)
					fileIsOpen = true
				// continue
				case CmdUnTune:
					log.Printf("WARN cmd.Type %v should not be called here", cmd.Type)
					// stopFfPlayAndLongmynd()
				case CmdEnableOffset:
					log.Printf("WARN cmd.Type %v should not be called here", cmd.Type)
					// enableOffset()
				case CmdDisableOffset:
					log.Printf("WARN cmd.Type %v should not be called here", cmd.Type)
					// disableOffset()
				default:
					log.Fatalf("FATAL cmd.Type was %v", cmd.Type)
				}
				// default:
			}
			fmt.Println(".")
			continue
		}

		select {
		case <-ctx.Done():
			stopFfPlayAndLongmynd()
			file.Close()
			log.Printf("CANCEL ----- lmClient 2 has cancelled")
			return
		case cmd := <-lmCmdChan:
			switch cmd.Type {
			case CmdTune:
				log.Printf("INFO ------ WILL TUNE")
				startLongmynd(cmd.FrequencyStr, cmd.SymbolRateStr)
			case CmdUnTune:
				log.Printf("INFO ------ WILL UNTUNE")
				stopFfPlayAndLongmynd()
			case CmdEnableOffset:
				enableOffset()
			case CmdDisableOffset:
				disableOffset()

			}
			// default:
		}

		rawStr, err := reader.ReadString(10) // delimited by char(10) == LF
		if err != nil {
			log.Printf("ERROR reading fifo: %v", err)
			liveData.reset()
			cacheData.reset()
			lmDataChan <- *liveData
			// time.Sleep(100 * time.Millisecond)
			continue
		}

		lmId, lmVal, err := idAndValFromString(rawStr)
		if err != nil {
			log.Printf("WARN Returned from idAndValFromString: %v", err)
			continue
		}

		switch lmId {
		case 1: // State
			liveData.id1_setState(lmVal)
			isLocked = liveData.State == kLocked
			if !isLocked { // if not locked, reset most status
				liveData.resetPartial()
				cacheData.reset()
				esPair.reset()
				agcPair.reset()
				lmDataChan <- *liveData
				// time.Sleep(5 * time.Millisecond)
				continue
			}
		// case 2: // LNA Gain - On devices that have LNA Amplifiers this represents the two gain sent as N, where n = (lna_gain<<5) | lna_vgo. Though not actually linear, n can be usefully treated as a single byte representing the gain of the amplifier
		// case 3: // Puncture Rate - During a search this is the pucture rate that is being trialled. When locked this is the pucture rate detected in the stream. Sent as a single value, n, where the pucture rate is n/(n+1)
		// case 4: // I Symbol Power - Measure of the current power being seen in the I symbols
		// case 5: // Q Symbol Power - Measure of the current power being seen in the Q symbols
		case 6: // Carrier Frequency - During a search this is the carrier frequency being trialled. When locked this is the Carrier Frequency detected in the stream. Sent in KHz
			liveData.id6_setFrequency(lmVal) //, offset)
		// case 7: // I Constellation - Single signed byte representing the voltage of a sampled I point
		// case 8: // Q Constellation - Single signed byte representing the voltage of a sampled Q point
		case 9: // Symbol Rate - During a search this is the symbol rate being trialled.  When locked this is the symbol rate detected in the stream
			liveData.id9_setSymbolRate(lmVal)
		// case 10: // Viterbi Error Rate - Viterbi correction rate as a percentage * 100
		// case 11: // BER - Bit Error Rate as a Percentage * 100
		case 12: // MER - Modulation Error Ratio in dB * 10
			liveData.id12_setDbMer(lmVal)
		case 13: // Service Provider - TS Service Provider Name
			liveData.id13_setProvider(lmVal)
		case 14: // Service Provider Service - TS Service Name
			liveData.id14_setService(lmVal)
		case 15: // Null Ratio - Ratio of Nulls in TS as percentage
			liveData.id15_setNullRatio(lmVal)
		case 16: // The PID numbers themselves are fairly arbitrary, will vary based on the transmitted signal and don't really mean anything in a single program multiplex.
			id16_setEsPid(lmVal)
		case 17: // ES TYPE - Elementary Stream Type (repeated as pair with 16 for each ES)
			liveData.id17_setEsType(lmVal)
		case 18: // MODCOD - Received Modulation & Coding Rate. See MODCOD Lookup Table below
			liveData.id18_setConstellationAndFecAndMargin(lmVal)
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
			liveData.id27_setDbmPower(lmVal)
		} // switch

		if isTuned && isLocked && !isPlaying {
			startFfplay()
		}
		if isTuned && !isLocked && isPlaying {
			stopFfplay()
		}
		if !isTuned && isPlaying {
			isLocked = false
			stopFfPlayAndLongmynd()
		}

		if isLocked {
			liveData.StatusMsg = fmt.Sprintf("%s : %s : %s", liveData.State, liveData.Provider, liveData.Service)
		} else {
			liveData.StatusMsg = liveData.State
		}

		if *liveData != *cacheData {
			lmDataChan <- *liveData
			*cacheData = *liveData
		}
	}
}

func enableOffset() {

}

func disableOffset() {

}

func (d *LmData_t) reset() {
	d.resetPartial()
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
}

var (
	lmExecCmd      *exec.Cmd
	fpExecCmd      *exec.Cmd
	ffPlayIsACtive bool // TODO: temp fix to prevent more than one ffplay instance
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

	const_ModeFecThreshold = map[string]float64{
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

	const_Agc1 = [...][2]int{
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

	const_Agc2 = [...][2]int{
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

type esPairStuct struct {
	waitingFor1stPid  bool
	waitingFor2ndPid  bool
	waitingFor1stType bool
	waitingFor2ndType bool
	the1stPidValue    string
	the1stTypeValue   string
	the2ndPidValue    string
	the2ndTypeValue   string
}

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

type agcPairStuct struct {
	waitingForAgc2 bool
	the1stAgcValue int
	the2ndAgcValue int
}

func (d *agcPairStuct) reset() {
	d.waitingForAgc2 = false
	d.the1stAgcValue = 0
	d.the2ndAgcValue = 0
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
	esPair    = esPairStuct{}
	agcPair   = new(agcPairStuct)
	liveData  = new(LmData_t)
	cacheData = new(LmData_t)
	isTuned   bool
	isPlaying bool
)

/***********************************************************
	functions called from the main switch statement
***********************************************************/

// State
func (d *LmData_t) id1_setState(stateStr string) {
	switch stateStr {
	case "0":
		d.State = kInitialising
	case "1":
		d.State = kSeaching
	case "2":
		d.State = kFoundHeaders
	case "3":
		d.State = kLocked
		d.Mode = kDVB_S
	case "4":
		d.State = kLocked
		d.Mode = kDVB_S2
	default:
		log.Printf("WARN Undefined status: %v", stateStr)
	}
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
}

// Service Provider - TS Service Provider Name
func (d *LmData_t) id13_setProvider(providerStr string) {
	if providerStr == "" {
		d.Provider = kDash
		return
	}
	d.Provider = providerStr
}

// Service Provider Service - TS Service Name
func (d *LmData_t) id14_setService(serviceStr string) {
	if serviceStr == "" {
		d.Service = kDash
		return
	}
	d.Service = serviceStr
}

// Null Ratio - Ratio of Nulls in TS as percentage
func (d *LmData_t) id15_setNullRatio(nullRatioStr string) {
	if nullRatioStr == "" {
		log.Printf("WARN Missing nullRatioStr")
		d.NullRatio = kDash
		return
	}
	d.NullRatio = nullRatioStr
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
	if esPair.waitingFor1stType {
		esPair.the1stTypeValue = esType
		esPair.waitingFor1stPid = false
		esPair.waitingFor2ndPid = true
		esPair.waitingFor1stType = false
		esPair.waitingFor2ndType = false
		d.PidPair1 = fmt.Sprintf("%v %v", esPair.the1stPidValue, esPair.the1stTypeValue) // beacon 257 27 = video
	}
	if esPair.waitingFor2ndType {
		esPair.the2ndTypeValue = esType
		esPair.waitingFor1stPid = true
		esPair.waitingFor2ndPid = false
		esPair.waitingFor1stType = false
		esPair.waitingFor2ndType = false
		d.PidPair2 = fmt.Sprintf("%v %v", esPair.the2ndPidValue, esPair.the2ndTypeValue) // beacon 258 3 = audio
		// NEWesPair.reset()
	}

	typ, err := strconv.Atoi(esType)
	if err != nil {
		log.Printf("WARN Failed to convert esType %v", err)
		return
	}
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
}

// AGC1 Gain - Gain value of AGC1 (0: Signal too weak, 65535: Signal too strong)
func id26_setDbmPower(agc1Str string) {
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

	d.DbmPower = fmt.Sprint(power)
	agcPair.reset()
}

/***********************************************************************
*
*	START AND STOP FUNCTIONS
*
************************************************************************/

func stopFfPlayAndLongmynd() {
	if isPlaying {
		stopFfplay()
	}
	if isTuned {
		stopLongmynd()
	}
}

// Start Longmynd with frequency and symbolrate
//
//	ie. /home/pi/q100/longmynd/longmynd -S 0.6 requestKHzStr symbolRateStr
func startLongmynd(frequency, symbolRate string) {
	// trim "10491.50 / 00" to "10491.50"
	frequencySplit := strings.SplitN(frequency, " ", 2)[0]
	requestedFrequency, err := strconv.ParseFloat(frequencySplit, 64)
	if err != nil {
		log.Fatalf("FATAL bad lmFrequency: %v", err)

	}
	requestKHz := (requestedFrequency * 1000) - config_LmOffset
	requestKHzStr := strconv.FormatFloat(requestKHz, 'f', 0, 64)
	log.Printf("INFO longmynd will start...")
	lmExecCmd = exec.Command("./longmynd", "-S", "0.6", requestKHzStr, symbolRate)
	lmExecCmd.Dir = config_LmFolder // ie. /home/pi/Q100/longmynd/
	if err = lmExecCmd.Start(); err != nil {
		log.Printf("ERROR failed to start longmynd: %v", err)
		return
	}
	log.Printf("INFO longmynd has started with f = %v", requestKHzStr)
	isTuned = true
}

// Stop Longmynd
func stopLongmynd() {
	if isTuned {
		log.Printf("INFO longmynd will stop...")
		lmExecCmd.Process.Kill()
		lmExecCmd.Process.Wait()
		cmd := exec.Command("/usr/bin/pkill", "longmynd")
		if err := cmd.Start(); err != nil {
			log.Printf("ERROR failed to stop longmynd: %v", err)
			return
		}
		cmd.Wait()
	}
	log.Printf("INFO longmynd has stopped")
	isTuned = false
}

// Start ffplay
//
//	ie. with position in frame buffer, fullscreen and volume
func startFfplay() {
	if !isPlaying && !ffPlayIsACtive {
		log.Printf("INFO ffplay will start...")
		fpExecCmd = exec.Command("/usr/bin/ffplay", "-left", "800", "-fs", "-volume", config_FpVolume, "-i", config_FpTsFifo)
		if err := fpExecCmd.Start(); err != nil {
			log.Printf("ERROR failed to start ffplay: %v", err)
			return
		}
		// cmd.Wait()
		log.Printf("INFO ffplay has started")
	}
	ffPlayIsACtive = true
	isPlaying = true
}

// Stop ffplay
func stopFfplay() {
	if isPlaying {
		log.Printf("INFO ffplay will stop...")
		fpExecCmd.Process.Kill()
		fpExecCmd.Process.Wait()
		cmd := exec.Command("/usr/bin/pkill", "ffplay")
		if err := cmd.Start(); err != nil {
			log.Printf("ERROR failed to stop ffplay: %v", err)
			return
		}
		cmd.Wait()
	}
	log.Printf("INFO ffplay has stppoed")
	ffPlayIsACtive = false
	isPlaying = false
}
