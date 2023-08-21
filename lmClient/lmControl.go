package lmClient

import (
	"os"
	"os/exec"
	"q100receiver/mylogger"
	"strconv"
	"strings"
	"syscall"
)

/****************************************************
*****************************************************
*** V 1 *********************************************
*****************************************************
*****************************************************/

func V1_killAll() bool {
	cmd := exec.Command("/usr/bin/killall", "-SIGINT", "ffplay", "pulseaudio", "longmynd") // "longmynd", "ffplay", "pulseaudio")
	_, _ = cmd.Output()
	isPlaying = false
	return false
}

func V1_startLongmynd(frequency, symbolRate string) bool {
	// trim "10491.50 / 00" to "10491.50"
	frequencySplit := strings.SplitN(frequency, " ", 2)[0]
	requestedFrequency, err := strconv.ParseFloat(frequencySplit, 64)
	if err != nil {
		mylogger.Fatal.Fatalf("bad lmFrequency: %v", err)
	}
	requestKHz := (requestedFrequency * 1000) - lmcfg.Offset
	requestKHzStr := strconv.FormatFloat(requestKHz, 'f', 0, 64)
	mylogger.Info.Printf("longmynd will start...")
	cmd := exec.Command(lmcfg.StartScript, lmcfg.Folder, lmcfg.Binary, requestKHzStr, symbolRate)
	mylogger.Info.Printf("exec: %v", cmd)
	_, err = cmd.Output()
	if err != nil {
		mylogger.Fatal.Fatalf("unable to start longmynd: %v", err)
		return false
	}
	mylogger.Info.Printf("longmynd has started")
	return true
}

func V1_stopLongmynd() bool {
	mylogger.Info.Printf("longmynd will stop...")
	cmd := exec.Command(lmcfg.StopScript)
	mylogger.Info.Printf("exec: %v", cmd.Args)
	_, err := cmd.Output()
	if err != nil {
		mylogger.Warn.Printf("unable to kill longmynd: %v", err)
		return false
	}
	mylogger.Info.Printf("longmynd has stopped")
	return false
}

func V1_startFfplay() bool {
	mylogger.Info.Printf("ffplay will start...")
	cmd := exec.Command(fpcfg.StartScript, fpcfg.Volume, fpcfg.TsFifo)
	mylogger.Info.Printf("exec: %v", cmd.Args)
	_, err := cmd.Output()
	if err != nil {
		mylogger.Fatal.Fatalf("unable to start ffplay: %v", err)
		return false
	}
	mylogger.Info.Printf("ffplay has started")
	return true
}

func V1_stopFfplay() bool {
	mylogger.Info.Printf("ffplay will stop...")
	cmd := exec.Command(fpcfg.StopScript)
	mylogger.Info.Printf("exec: %v", cmd)
	_, err := cmd.Output()
	if err != nil {
		mylogger.Warn.Printf("unable to stop ffplay: %v", err)
		return false
	}
	mylogger.Info.Printf("ffplay has stopped")
	isPlaying = false
	return false
}

/****************************************************
*****************************************************
*** V 2 *********************************************
*****************************************************
*****************************************************/

func V2_killAll() {
	if isPlaying {
		V2_stopFfplay()
	}
	if isTuned {
		V2_stopLongmynd()
	}
}

func V2_startLongmynd(frequency, symbolRate string) {
	// trim "10491.50 / 00" to "10491.50"
	frequencySplit := strings.SplitN(frequency, " ", 2)[0]
	requestedFrequency, err := strconv.ParseFloat(frequencySplit, 64)
	if err != nil {
		mylogger.Fatal.Fatalf("bad lmFrequency: %v", err)
		return
	}
	requestKHz := (requestedFrequency * 1000) - lmcfg.Offset
	requestKHzStr := strconv.FormatFloat(requestKHz, 'f', 0, 64)
	mylogger.Info.Printf("longmynd will start...")
	_, err = exec.Command(lmcfg.StartScript, lmcfg.Folder, lmcfg.Binary, requestKHzStr, symbolRate).Output()
	if err != nil {
		mylogger.Error.Printf("failed to start longmynd: %v", err)
		return
	}
	mylogger.Info.Printf("longmynd has started")
	isTuned = true
}

func V2_stopLongmynd() {
	if isTuned {
		mylogger.Info.Printf("longmynd will stop...")
		_, err := exec.Command("/usr/bin/pkill", "longmynd").Output()
		if err != nil {
			mylogger.Error.Printf("failed to stop longmynd: %v", err)
			return
		}
		mylogger.Info.Printf("longmynd has stopped")
	}
	isTuned = false
}

// export DISPLAY=:0
// /usr/bin/ffplay -left 800 -fs -volume "$1" -i "$2" > /dev/null 2>&1 &
var ffPlayPID int

func V2bad_startFfplay() {
	if !isPlaying {
		func() {
			mylogger.Info.Printf("ffplay will start...")
			cmd := exec.Command("/usr/bin/ffplay", "-left", "800", "-fs", "-i", fpcfg.TsFifo)
			err := cmd.Start()
			if err != nil {
				mylogger.Error.Printf("failed to start ffplay: %v", err)
				return
			}
			ffPlayPID = cmd.Process.Pid
			mylogger.Info.Printf("ffplay has started %v", ffPlayPID)
			isPlaying = true
			_, err = cmd.Process.Wait()
			if err != nil {
				mylogger.Error.Printf("failed to wait ffplay: %v", err)
			}
			isPlaying = false
		}()
	}
}

func V2_startFfplay() {
	if !isPlaying {
		mylogger.Info.Printf("ffplay will start...")
		_, err := exec.Command(fpcfg.StartScript, fpcfg.Volume, fpcfg.TsFifo).Output()
		if err != nil {
			mylogger.Error.Printf("failed to start ffplay: %v", err)
			return
		}
		mylogger.Info.Printf("ffplay has started")
	}
	isPlaying = true
}

func V2_stopFfplay() {
	if isPlaying {
		mylogger.Info.Printf("ffplay will stop...")
		// err := syscall.Kill(ffPlayPID, syscall.SIGINT) /// syscall.SIGINT, syscall.SIGTERM
		_, err := exec.Command("/usr/bin/pkill", "ffplay").Output()
		if err != nil {
			mylogger.Error.Printf("failed to stop ffplay: %v", err)
			return
		}
		// apparently pulseaudio is allreaady running
		// _, err = exec.Command("/usr/bin/pkill", "pulseaudio").Output()
		// if err != nil {
		// 	mylogger.Error.Printf("failed to stop pulseaudio: %v", err)
		// 	return
		// }
	}
	mylogger.Info.Printf("ffplay has stppoed")
	isPlaying = false
}

/****************************************************
*****************************************************
*** V 3 *********************************************
*****************************************************
*****************************************************/

func V3_killAll() {
	//
}

func V3_startLongmynd() { // TODO: return isTuned etc.
	mylogger.Info.Printf("longmynd will start...")

	// trim "10491.50 / 00" to "10491.50"
	freqeuncy := strings.SplitN(withFrequency, " ", 2)[0]
	requestedFrequency, err := strconv.ParseFloat(freqeuncy, 64)
	if err != nil {
		mylogger.Warn.Printf("bad lmFrequency: %v", err)
		return
	}
	requestKHz := (requestedFrequency * 1000) - lmcfg.Offset
	requestKHzStr := strconv.FormatFloat(requestKHz, 'f', 0, 64)
	var args []string
	// args = append(args, lmcfg.Binary)
	args = append(args, "-S 0.6")
	args = append(args, requestKHzStr)
	args = append(args, withSysmbolRate)
	var procAttr os.ProcAttr
	procAttr.Dir = lmcfg.Folder
	// procAttr.Files = []*os.File{os.Stdout, os.Stderr}
	p, err := os.StartProcess(lmcfg.Binary, args, &procAttr)
	if err != nil {
		lmPid = 0
		mylogger.Warn.Printf("longmynd failed to start: %v", err)
		return
	}
	lmPid = p.Pid
	mylogger.Info.Printf("longmynd has started with PID: %v", lmPid)
}

func V3_stopLongmynd() {
	if lmPid == 0 {
		return
	}
	mylogger.Info.Printf("longmynd will stop...")
	err := syscall.Kill(lmPid, syscall.SIGINT)
	if err != nil {
		lmPid = 0
		mylogger.Warn.Printf("unable to kill longmynd: %v", err)
		return
	}
	lmPid = 0
	mylogger.Info.Printf("longmynd has stopped")
}

func V3_startFfplay() {
	// if ffPlayPid != 0 {
	// 	return
	// }
	mylogger.Info.Printf("ffplay will start...")
	// time.Sleep(time.Second)
	// // cmd := exec.Command(fpcfg.Binary, "-left 800", "-fs", "-volume "+fpcfg.Volume, "-i "+fpcfg.TsFifo)
	// cmd := exec.Command(fpcfg.Binary, "-left", "800", "-fs", "-volume", fpcfg.Volume, "-i ", fpcfg.TsFifo)
	// cmd.Env = append(os.Environ(),
	// 	"DISPLAY=:0",
	// )
	// if err := cmd.Run(); err != nil {
	// 	mylogger.Fatal.Fatalf(": %v", err)
	// }
	// something, err := cmd.Output()
	// if err != nil {
	// 	mylogger.Warn.Printf("failed to start ffplay: %v", err)
	// }
	// mylogger.Info.Printf("ffplay has started %v", something)
	// return

	var args []string
	// args = append(args, "ffplay")
	args = append(args, "-left 800")
	args = append(args, "-fs")
	args = append(args, "-volume "+fpcfg.Volume)
	args = append(args, "-i "+fpcfg.TsFifo)

	// export DISPLAY=:0

	var procAttr os.ProcAttr
	// fmt.Println(os.Environ())
	procAttr.Env = append(os.Environ(),
		"DISPLAY=:0",
	)
	// procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	procAttr.Files = []*os.File{os.Stdout, os.Stderr}
	p, err := os.StartProcess(fpcfg.Binary, args, &procAttr)
	if err != nil {
		ffPlayPid = 0
		mylogger.Warn.Printf("failed to start ffplay: %v", err)
		return
	}
	ffPlayPid = p.Pid
	mylogger.Info.Printf("ffplay has started with PID: %v ARGS: %v", ffPlayPid, args)
}

func V3_stopFfplay() {
	mylogger.Info.Printf("ffplay will stop...")

	cmd := exec.Command("/usr/bin/killall", "-SIGINT", "ffplay", "pulseaudio")
	_, _ = cmd.Output()

	// if ffPlayPid == 0 {
	// 	return
	// }
	// mylogger.Info.Printf("ffplay will stop...")
	// err := syscall.Kill(ffPlayPid, syscall.SIGINT)
	// if err != nil {
	// 	mylogger.Warn.Printf("unable to kill ffplay: %v", err)
	// 	return
	// }
	// ffPlayPid = 0
	mylogger.Info.Printf("ffplay has stopped")
}

/****************************************************
*****************************************************
*** V 4 *********************************************
*****************************************************
*****************************************************/
