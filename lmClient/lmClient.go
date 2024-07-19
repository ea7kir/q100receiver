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
	LmCmd_t struct {
		Type          int
		FrequencyStr  string
		FrequencyKHz  float64
		SymbolRateStr string
	}

	LmData_t struct {
		changed bool
		// isLocked      bool
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
	dependant             = lmDependants_t{}
)

///////////////////////////////////////////////////////////////////////////////////////////

// Reads the longmynd status fifo and translates to formated strings.
//
//	The results are sent to a channel of type LongmyndData. When no valid signal is being
//	received, the LongmyndData fileds will be filled with default values - normally a dash.
//
// func ReadLonmyndStatus(ctx context.Context, lmc LmConfig_t, fpc FpConfig_t, ch chan LongmyndData) {
func ReadLonmyndStatus(ctx context.Context, lmCmdChan <-chan LmCmd_t, lmDataChan chan<- LmData_t) {

	liveData := LmData_t{}

	liveData.reset()
	// cacheData.reset()
	esPair.reset()

	isLocked := false

	lmDataChan <- liveData

	// var fifo *os.File = nil
	var reader *bufio.Reader = nil
	// var fifoIsOpen = false

	for {
		if !dependant.fifoIsOpen {
			select {
			case <-ctx.Done():
				dependant.stopFfPlayAndLongmynd()
				// fifo.Close()
				log.Printf("CANCEL ----- lmClient 1 has cancelled")
				return
			case cmd := <-lmCmdChan:
				switch cmd.Type {
				case CmdTune:
					log.Printf("INFO ------ WILL TUNE for the first time")
					dependant.startLongmynd(cmd.FrequencyStr, cmd.SymbolRateStr)
					// MOVED TO startLongmynd
					var err error
					dependant.fifo, err = os.OpenFile(config_LmStatusFifo, os.O_RDONLY, os.ModeNamedPipe)
					if err != nil {
						log.Fatalf("FATAL Failed to open '%v' fifo %v: ", config_LmStatusFifo, err)
					}
					reader = bufio.NewReader(dependant.fifo)
					dependant.fifoIsOpen = true
					continue
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
			// log.Println("TEMP .")
			continue
		}

		select {
		case <-ctx.Done():
			dependant.stopFfPlayAndLongmynd()
			dependant.fifo.Close() // MOVED TO stopLongmynd
			log.Printf("CANCEL ----- lmClient 2 has cancelled")
			return
		case cmd := <-lmCmdChan:
			switch cmd.Type {
			case CmdTune:
				log.Printf("INFO ------ WILL TUNE")
				dependant.startLongmynd(cmd.FrequencyStr, cmd.SymbolRateStr)
			case CmdUnTune:
				log.Printf("INFO ------ WILL UNTUNE")
				dependant.stopFfPlayAndLongmynd()
			case CmdEnableOffset:
				enableOffset()
			case CmdDisableOffset:
				disableOffset()

			}
		default:
		}
		// log.Printf("TEMP in lmClient loop")
		rawStr, err := reader.ReadString(10) // delimited by char(10) == LF
		if err != nil {
			log.Printf("ERROR reading fifo: %v", err)
			liveData.reset()
			// cacheData.reset()
			lmDataChan <- liveData
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
				// cacheData.reset()
				esPair.reset()
				agcPair.reset()
				lmDataChan <- liveData
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
			liveData.id16_setEsPid(lmVal)
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
			liveData.id26_setDbmPower(lmVal)
		case 27: // AGC2 Gain - Gain value of AGC2 (0: Minimum Gain, 65535: Maximum Gain)
			liveData.id27_setDbmPower(lmVal)
		} // switch

		if dependant.isTuned && isLocked && !dependant.isPlaying {
			dependant.startFfplay()
		}
		if dependant.isTuned && !isLocked && dependant.isPlaying {
			dependant.stopFfplay()
		}
		if !dependant.isTuned && dependant.isPlaying {
			isLocked = false
			dependant.stopFfPlayAndLongmynd()
		}

		if isLocked {
			liveData.StatusMsg = fmt.Sprintf("%s : %s : %s", liveData.State, liveData.Provider, liveData.Service)
		} else {
			liveData.StatusMsg = liveData.State
		}

		if liveData.changed {
			lmDataChan <- liveData
			liveData.changed = false
		}
	}
}

func enableOffset() {

}

func disableOffset() {

}

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

// const (
// 	kDash         = "-"
// 	kInitialising = "Initialising"
// 	kSeaching     = "Seaching"
// 	kFoundHeaders = "Found Headers"
// 	kLocked       = "Locked"
// 	kDVB_S        = "DVB-S"
// 	kDVB_S2       = "DVB-S2"
// )

// /***********************************************************
// 	functions called from the main switch statement
// ***********************************************************/
