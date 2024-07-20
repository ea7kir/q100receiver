package lmClient

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

/***********************************************************************
*
*	START AND STOP FUNCTIONS
*
************************************************************************/

type (
	lmDependants_t struct {
		isPlaying      bool
		isTuned        bool
		ffPlayIsACtive bool
		lmExecCmd      *exec.Cmd
		fpExecCmd      *exec.Cmd
		fifo           *os.File
	}
)

func (d *lmDependants_t) stopFfPlayAndLongmynd() {
	if d.isPlaying {
		d.stopFfplay()
	}
	if d.isTuned {
		d.stopLongmynd()
	}
}

// Start Longmynd with frequency and symbolrate
//
//	ie. /home/pi/q100/longmynd/longmynd -S 0.6 requestKHzStr symbolRateStr
func (d *lmDependants_t) startLongmynd(frequency, symbolRate string) {
	// trim "10491.50 / 00" to "10491.50"
	frequencySplit := strings.SplitN(frequency, " ", 2)[0]
	requestedFrequency, err := strconv.ParseFloat(frequencySplit, 64)
	if err != nil {
		log.Fatalf("FATAL bad lmFrequency: %v", err)

	}
	requestKHz := (requestedFrequency * 1000) - config_LmOffset
	requestKHzStr := strconv.FormatFloat(requestKHz, 'f', 0, 64)
	log.Printf("INFO longmynd will start...")
	d.lmExecCmd = exec.Command("./longmynd", "-S", "0.6", requestKHzStr, symbolRate)
	d.lmExecCmd.Dir = config_LmFolder // ie. /home/pi/Q100/longmynd/
	if err = d.lmExecCmd.Start(); err != nil {
		log.Printf("ERROR failed to start longmynd: %v", err)
		return
	}
	log.Printf("INFO longmynd has started with f = %v", requestKHzStr)

	d.fifo, err = os.OpenFile(config_LmStatusFifo, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		log.Fatalf("FATAL Failed to open '%v' fifo %v: ", config_LmStatusFifo, err)
	}
	log.Printf("INFO fifo is open %v", d.fifo.Name())
	d.isTuned = true
}

// Stop Longmynd
func (d *lmDependants_t) stopLongmynd() {
	// if d.isTuned {
	log.Printf("INFO longmynd will stop...")
	d.lmExecCmd.Process.Kill()
	d.lmExecCmd.Process.Wait()
	cmd := exec.Command("/usr/bin/pkill", "longmynd")
	if err := cmd.Start(); err != nil {
		log.Printf("ERROR failed to stop longmynd: %v", err)
		return
	}
	cmd.Wait()
	// }
	log.Printf("INFO longmynd has stopped")
	d.isTuned = false
	d.fifo.Close() // TODO: should this higher up ?
}

// Start ffplay
//
//	ie. with position in frame buffer, fullscreen and volume
func (d *lmDependants_t) startFfplay() {
	if !d.isPlaying && !d.ffPlayIsACtive {
		log.Printf("INFO ffplay will start...")
		d.fpExecCmd = exec.Command("/usr/bin/ffplay", "-left", "800", "-fs", "-volume", config_FpVolume, "-i", config_FpTsFifo)
		if err := d.fpExecCmd.Start(); err != nil {
			log.Printf("ERROR failed to start ffplay: %v", err)
			return
		}
		// cmd.Wait()
		log.Printf("INFO ffplay has started")
	}
	d.ffPlayIsACtive = true
	d.isPlaying = true
}

// Stop ffplay
func (d *lmDependants_t) stopFfplay() {
	if d.isPlaying {
		log.Printf("INFO ffplay will stop...")
		d.fpExecCmd.Process.Kill()
		d.fpExecCmd.Process.Wait()
		cmd := exec.Command("/usr/bin/pkill", "ffplay")
		if err := cmd.Start(); err != nil {
			log.Printf("ERROR failed to stop ffplay: %v", err)
			return
		}
		cmd.Wait()
	}
	log.Printf("INFO ffplay has stppoed")
	d.ffPlayIsACtive = false
	d.isPlaying = false
}
