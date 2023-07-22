package lmClient

import "q100receiver/logger"

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

var (
	lmcfg *LmConfig
	fpcfg *FpConfig
)

func Intitialize(lmc LmConfig, fpc FpConfig, ch chan LongmyndData) {
	lmcfg = &lmc
	fpcfg = &fpc
	lmChannel = ch

	killAll()
	go readLongmynd(lmcfg.StatusFifo, lmcfg.Offset, lmChannel)
}

func Stop() {
	logger.Info("LmReader will stop...")
	// TODO: implement a better way to stop longmynd and ffplay
	killAll()
	logger.Info("LmReader has stopped")
}

func Tune(frequency, sysmbolRate string) {
	logger.Info("------ WILL TUNE")
	isTuned = startLongmynd(frequency, sysmbolRate) // TODO: pass arguments
	// isTuned = true
}

func UnTune() {
	logger.Info("------ WILL UNTUNE")
	// isTuned = stopLongmynd()
	killAll()
	isTuned = false
	// stopFfplay()
}
