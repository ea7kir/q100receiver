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
	"strconv"
	"strings"
	"time"
)

const (
	config_LmBaseFolder = "/home/pi/Q100/"

	config_LmFolder = config_LmBaseFolder + "longmynd/"
	// config_LmBinary          = config_LmBaseFolder + "longmynd/longmynd"
	config_LmStatusFifo      = config_LmBaseFolder + "longmynd/longmynd_main_status"
	config_LmOffset_Received = float64(9750000 - 52) // only the displayed frequency
	config_LmOffset_Reqested = float64(9750000 + 0)

	config_FpTsFifo = config_LmBaseFolder + "longmynd/longmynd_main_ts"
	// config_FpBinary = "/usr/bin/ffplay"
	config_FpVolume = "100"

	CmdTune   = 1
	CmdUnTune = 2
	// CmdToggleCalibrate = 3
	// CmdEnableOffset  = 3
	// CmdDisableOffset = 4
)

type (
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
		changed       bool
		Locked        bool
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

///////////////////////////////////////////////////////////////////////////////////////////

// Reads the longmynd status fifo and translates to formated strings.
//
//	The results are sent to a channel of type LongmyndData. When no valid signal is being
//	received, the LongmyndData fileds will be filled with default values - normally a dash.
//
// func ReadLonmyndStatus(ctx context.Context, lmc LmConfig_t, fpc FpConfig_t, ch chan LongmyndData) {
func ReadLonmyndStatus(ctx context.Context, lmCmdChan <-chan LmCmd_t, lmDataChan chan<- LmData_t) {

	liveData := LmData_t{}
	dependant := lmDependants_t{}

	liveData.reset()
	lmDataChan <- liveData

	var reader *bufio.Reader = nil

	for {
		select {
		case <-ctx.Done():
			dependant.stopFfPlayAndLongmynd()
			log.Printf("CANCEL ----- lmClient has cancelled")
			return
		case cmd := <-lmCmdChan:
			switch cmd.Type {
			case CmdTune:
				log.Printf("INFO ------ WILL TUNE")
				dependant.startLongmynd(cmd.FrequencyStr, cmd.SymbolRateStr)
				reader = bufio.NewReader(dependant.fifo)
			case CmdUnTune:
				log.Printf("INFO ------ WILL UNTUNE")
				dependant.stopFfPlayAndLongmynd()
				// case CmdToggleCalibrate:
				// TODO: implement
			}
		default:
		}

		if !dependant.isTuned {
			time.Sleep(time.Microsecond * 50)
			liveData.reset()
			lmDataChan <- liveData
			continue
		}

		rawStr, err := reader.ReadString(10) // delimited by char(10) == LF
		if err != nil {
			log.Printf("ERROR reading fifo: %v", err)
			liveData.reset()
			lmDataChan <- liveData
			time.Sleep(time.Millisecond * 50)
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
			if !liveData.Locked { // if not locked, reset most status
				liveData.resetPartial()
				lmDataChan <- liveData
				// time.Sleep(time.Microsecond * 50)
				continue
			}
		// case 2: // LNA Gain - On devices that have LNA Amplifiers this represents the two gain sent as N, where n = (lna_gain<<5) | lna_vgo. Though not actually linear, n can be usefully treated as a single byte representing the gain of the amplifier
		// case 3: // Puncture Rate - During a search this is the pucture rate that is being trialled. When locked this is the pucture rate detected in the stream. Sent as a single value, n, where the pucture rate is n/(n+1)
		// case 4: // I Symbol Power - Measure of the current power being seen in the I symbols
		// case 5: // Q Symbol Power - Measure of the current power being seen in the Q symbols
		case 6: // Carrier Frequency - During a search this is the carrier frequency being trialled. When locked this is the Carrier Frequency detected in the stream. Sent in KHz
			liveData.id6_setFrequency(lmVal, dependant.requestKHz) //, offset)
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

		// TODO: the follwoing 4 if statement should be in a function in lmDependants.go

		if dependant.isTuned && liveData.Locked && !dependant.isPlaying {
			dependant.startFfplay()
		}
		if dependant.isTuned && !liveData.Locked && dependant.isPlaying {
			dependant.stopFfplay()
		}
		if !dependant.isTuned && dependant.isPlaying {
			liveData.Locked = false
			dependant.stopFfPlayAndLongmynd()
		}

		if liveData.Locked {
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
