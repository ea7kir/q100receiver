/*
 *  Q-100 Receiver
 *  Copyright (c) 2023 Michael Naylor EA7KIR (https://michaelnaylor.es)
 */

package main

import (
	"context"
	"flag"
	"image"
	"image/color"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"q100receiver/lmClient"
	"q100receiver/rxControl"
	"q100receiver/spClient"
	"syscall"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/ajstarks/giocanvas"
	"golang.org/x/image/colornames"
)

// local data
var (
	rxCmdChan  = make(chan rxControl.RxCmd_t)
	rxData     = rxControl.RxData_t{}
	rxDataChan = make(chan rxControl.RxData_t)
	spData     = spClient.SpData_t{}
	spDataChan = make(chan spClient.SpData_t, 1)
	lmData     = lmClient.LmData_t{}
	lmDataChan = make(chan lmClient.LmData_t)
)

func main() {
	time.Sleep(3 * time.Second)

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Printf("INFO ----- q100receiver Opened -----")

	var shutdown bool
	flag.BoolVar(&shutdown, "shutdown", false, "close and poweroff")
	flag.Parse()
	// fmt.Println("shudown: ", shutdown)

	ctx, cancel := context.WithCancel(context.Background())

	go spClient.ReadSpectrumServer(ctx, spDataChan)
	go rxControl.HandleCommands(ctx, rxCmdChan, rxDataChan, lmDataChan)

	go func() {
		const WINDOW_MANAGER = 2 // 1 = X!!, 2 = Wayfire, = Labwc
		switch WINDOW_MANAGER {
		case 1: // X11
			os.Setenv("XDG_RUNTIME_DIR", "/run/user/1000") // TODO: is 1000 corrrect?
			os.Setenv("DISPLAY", ":0")                     // required for X11. Compile wit: go build --tags nowayland .
		case 2: // Wayfire
			os.Setenv("XDG_RUNTIME_DIR", "/run/user/1000") // TODO: is 1000 corrrect?
			os.Setenv("WAYLAND_DISPLAY", "wayland-1")      // required for wayland. Compile with: go build --tags nox11 .
		case 3: // Labwc
			os.Setenv("XDG_RUNTIME_DIR", "/run/user/1000") // TODO: is 1000 corrrect?
			os.Setenv("WAYLAND_DISPLAY", "wayland-0")      // required for Labwc. Compile with: go build --tags nox11 .
		}

		app.Size(800, 480) // I don't know if this is help in any way
		var w app.Window
		w.Option(app.Fullscreen.Option())

		if err := loop(&w); err != nil {
			log.Fatalf("FATAL failed to start loop: %v", err)
		}

		cancel()
		// log.Printf("CANCEL IN MAIN ----- cancel() called")
		// allow time to cancel all functions
		time.Sleep(time.Second * 4)

		// TODO: control this with a flag
		if shutdown {
			log.Printf("INFO ----- q100receiver will poweroff -----")
			time.Sleep(1 * time.Second)
			cmd := exec.Command("sudo", "poweroff")
			if err := cmd.Start(); err != nil {
				log.Fatalf("FATAL failed to poweroff: %v", err)
			}
			cmd.Wait()
		}

		log.Printf("INFO ----- q100receiver Closed -----")
		// log.Close()
		os.Exit(0)
	}()

	app.Main()
}

func loop(w *app.Window) error {
	// ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	// defer stop()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	// defer signal.Stop(quit)

	ui := UI{
		// th: material.NewTheme(gofont.Collection()),
		th: material.NewTheme(),
	}
	// without this, the font sizes are inconsistent
	// Chris says keep using the original font
	ui.th.Shaper = text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))

	var ops op.Ops
	// Capture the context done channel in a variable so that we can nil it
	// out after it closes and prevent its select case from firing again.
	// done := ctx.Done()

	for {
		select {
		// case <-ctx.Done():
		case <-interrupt:
			// When the context cancels, assign the done channel to nil to
			// prevent it from firing over and over.
			interrupt = nil
			w.Perform(system.ActionClose)
		case rxData = <-rxDataChan:
			// log.Printf("TEMP got rxData")
			w.Invalidate()
		case lmData = <-lmDataChan:
			// log.Printf("TEMP got lmData")
			w.Invalidate()
		case spData = <-spDataChan:
			w.Invalidate()
		}

		switch event := w.Event().(type) {
		case app.DestroyEvent:
			return event.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, event)

			switch {
			case ui.about.Clicked(gtx):
				showAboutBox()
			case ui.shutdown.Clicked(gtx):
				interrupt <- syscall.SIGINT
				// w.Perform(system.ActionClose)
			case ui.decBand.Clicked(gtx):
				rxCmdChan <- rxControl.CmdDecBand
			case ui.incBand.Clicked(gtx):
				rxCmdChan <- rxControl.CmdIncBand
			case ui.decSymbolRate.Clicked(gtx):
				rxCmdChan <- rxControl.CmdDecSymbolRate
			case ui.incSymbolRate.Clicked(gtx):
				rxCmdChan <- rxControl.CmdIncSymbolRate
			case ui.decFrequency.Clicked(gtx):
				rxCmdChan <- rxControl.CmdDecFrequency
			case ui.incFrequency.Clicked(gtx):
				rxCmdChan <- rxControl.CmdIncFrequency
			case ui.tune.Clicked(gtx):
				rxCmdChan <- rxControl.CmdTune
			case ui.stream.Clicked(gtx):
				rxCmdChan <- rxControl.CmdStream
			}

			paint.Fill(gtx.Ops, q100color.screenGrey)
			ui.layoutFlexes(gtx)
			event.Frame(gtx.Ops)
		}
	}
}

// custom color scheme
var q100color = struct {
	screenGrey                               color.NRGBA
	labelWhite, labelOrange                  color.NRGBA
	buttonGrey, buttonGreen, buttonRed       color.NRGBA
	gfxBgd, gfxGreen, gfxGraticule, gfxLabel color.NRGBA
	gfxBeacon, gfxMarker                     color.NRGBA
}{
	// see: https://pkg.go.dev/golang.org/x/image/colornames
	// but maybe I should just create my own colors
	screenGrey:   color.NRGBA{R: 16, G: 16, B: 16, A: 255}, // no LightBlack
	labelWhite:   color.NRGBA(colornames.White),
	labelOrange:  color.NRGBA(colornames.Darkorange),       // or Orange or Darkorange or Gold
	buttonGrey:   color.NRGBA{R: 32, G: 32, B: 32, A: 255}, // DarkGrey is too light
	buttonGreen:  color.NRGBA(colornames.Green),
	buttonRed:    color.NRGBA(colornames.Red),
	gfxBgd:       color.NRGBA(colornames.Black),
	gfxGreen:     color.NRGBA(colornames.Green),
	gfxBeacon:    color.NRGBA(colornames.Red),
	gfxMarker:    color.NRGBA{R: 20, G: 20, B: 20, A: 255},
	gfxGraticule: color.NRGBA(colornames.Darkgray),
	gfxLabel:     color.NRGBA{R: 32, G: 32, B: 32, A: 255}, // DarkGrey is too light
}

// define all buttons
type UI struct {
	about, shutdown              widget.Clickable
	decBand, incBand             widget.Clickable
	decSymbolRate, incSymbolRate widget.Clickable
	decFrequency, incFrequency   widget.Clickable
	tune, stream                 widget.Clickable
	th                           *material.Theme
}

// makes the code more readable
type (
	C = layout.Context
	D = layout.Dimensions
)

/*********************************************************************************

[ [ button ]  [ label_____________________________________________ ]  [ button ] ]

[ [ ------------------------------- spectrum --------------------------------- ] ]

[    [ button label button ]  [ button label button ]  [ button label button ]   ]

[ [ label__  label__ ]   [ label__  label__ ]   [ label__  label__ ]  [ button ] ]
[ [ label__  label__ ]   [ label__  label__ ]   [ label__  label__ ]             ]
[ [ label__  label__ ]   [ label__  label__ ]   [ label__  label__ ]             ]
[ [ label__  label__ ]   [ label__  label__ ]   [ label__  label__ ]  [ button ] ]

*********************************************************************************/

// Returns an About box
func showAboutBox() {
	// TODO: implement an about box
}

// Return a customisable button
func (ui *UI) q100_Button(gtx C, button *widget.Clickable, label string, btnActive bool, btnActiveColor color.NRGBA) D {
	inset := layout.Inset{
		Top:    2,
		Bottom: 2,
		Left:   4,
		Right:  4,
	}

	btn := material.Button(ui.th, button, label)
	if btnActive {
		btn.Background = btnActiveColor
	} else {
		btn.Background = q100color.buttonGrey
	}
	btn.Color = q100color.labelWhite
	return inset.Layout(gtx, btn.Layout)
}

// Returns a customisable label
func (ui *UI) q100_Label(gtx C, label string, txtColor color.NRGBA) D {
	inset := layout.Inset{
		Top:    2,
		Bottom: 2,
		Left:   4,
		Right:  4,
	}

	lbl := material.Body1(ui.th, label)
	lbl.Color = txtColor
	return inset.Layout(gtx, lbl.Layout)
}

// Returns 1 row of 2 buttons and a label for About, Status and Shutdown
func (ui *UI) q100_TopStatusRow(gtx C) D {
	const btnWidth = 50
	inset := layout.Inset{
		Top:    2,
		Bottom: 2,
		Left:   4,
		Right:  4,
	}

	return layout.Flex{
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(btnWidth)
				return ui.q100_Button(gtx, &ui.about, "Q0-100 Receiver", false, q100color.buttonGrey)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			return ui.q100_Label(gtx, lmData.StatusMsg, q100color.labelOrange)
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(btnWidth)
				return ui.q100_Button(gtx, &ui.shutdown, "Shutdown", false, q100color.buttonGrey)
			})
		}),
	)
}

// Returns a single Selector_t as [ button label button ]
func (ui *UI) q100_Selector(gtx C, dec, inc *widget.Clickable, value string, btnWidth, lblWidth unit.Dp) D {
	inset := layout.Inset{
		Top:    2,
		Bottom: 2,
		Left:   4,
		Right:  4,
	}

	return layout.Flex{
		Axis: layout.Horizontal,
		// Spacing: layout.SpaceBetween,
		Alignment: layout.Middle, // Chris
		// WeightSum: 0.3,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(btnWidth)
				return ui.q100_Button(gtx, dec, "<", false, q100color.buttonGrey)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(lblWidth)
				return ui.q100_Label(gtx, value, q100color.labelOrange)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(btnWidth)
				return ui.q100_Button(gtx, inc, ">", false, q100color.buttonGrey)
			})
		}),
	)
}

// Returns 1 row of 3 Selectors for Band SymbolRate and Frequency
func (ui *UI) q100_MainTuningRow(gtx C) D {
	const btnWidth = 50

	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceEvenly,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return ui.q100_Selector(gtx, &ui.decBand, &ui.incBand, rxData.CurBand, btnWidth, 100)
		}),
		layout.Rigid(func(gtx C) D {
			return ui.q100_Selector(gtx, &ui.decSymbolRate, &ui.incSymbolRate, rxData.CurSymbolRate, btnWidth, 50)
		}),
		layout.Rigid(func(gtx C) D {
			return ui.q100_Selector(gtx, &ui.decFrequency, &ui.incFrequency, rxData.CurFrequency, btnWidth, 100)
		}),
	)
}

// Returns the Spectrum display
//
// see: github.com/ajstarks/giocanvas for docs
func (ui *UI) q100_SpectrumDisplay(gtx C) D {
	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceSides,
	}.Layout(gtx,
		layout.Rigid(
			func(gtx layout.Context) layout.Dimensions {
				canvas := giocanvas.Canvas{
					Width:   float32(788), //gtx.Constraints.Max.X), //float32(width),  //float32(gtx.Constraints.Max.X),
					Height:  float32(250), //float32(hieght), //float32(500),
					Context: gtx,
					Theme:   ui.th,
				}
				// fmt("  Canvas: %#v\n", canvas.Context.Constraints)

				canvas.Background(q100color.gfxBgd)
				// tuning marker
				canvas.Rect(rxData.MarkerCentre, 50, rxData.MarkerWidth, 100, q100color.gfxMarker)
				// polygon
				canvas.Polygon(spClient.Xp, spData.Yp, q100color.gfxGreen)
				// graticule
				const fyBase float32 = 3
				const fyInc float32 = 5.88235
				fy := fyBase
				for y := 0; y < 17; y++ {
					switch y {
					case 15:
						canvas.Text(1, fy, 1.5, "15dB", q100color.gfxLabel)
						canvas.HLine(5, fy, 94, 0.01, q100color.gfxGraticule)
					case 10:
						canvas.Text(1, fy, 1.5, "10dB", q100color.gfxLabel)
						canvas.HLine(5, fy, 94, 0.01, q100color.gfxGraticule)
					case 5:
						canvas.Text(1, fy, 1.5, "5dB", q100color.gfxLabel)
						canvas.HLine(5, fy, 94, 0.01, q100color.gfxGraticule)
					default:
						canvas.HLine(5, fy, 94, 0.005, q100color.gfxGraticule)
					}
					fy += fyInc
				}
				// beacon level
				canvas.HLine(5, spData.BeaconLevel, 94, 0.03, q100color.gfxBeacon)

				return layout.Dimensions{
					Size: image.Point{X: int(canvas.Width), Y: int(canvas.Height)},
				}
			},
		),
	)
}

// returns [ label__  label__ ]
func (ui *UI) q100_LabelValue(gtx C, label, value string) D {
	const lblWidth = 105
	const valWidth = 110
	inset := layout.Inset{
		Top:    2,
		Bottom: 2,
		Left:   4,
		Right:  4,
	}

	return layout.Flex{
		Axis: layout.Horizontal,
		// Spacing: layout.SpaceEnd,
		// Alignment: layout.Middle,
		// WeightSum: 0.3,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(lblWidth)
				gtx.Constraints.Max.X = gtx.Dp(lblWidth)
				return ui.q100_Label(gtx, label, q100color.labelWhite)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(valWidth)
				gtx.Constraints.Max.X = gtx.Dp(valWidth)
				return ui.q100_Label(gtx, value, q100color.labelOrange)
			})
		}),
	)
}

// returns a column of 4 rows of [label__  label__]
func (ui *UI) q100_Column4Rows(gtx C, name, value [4]string) D {
	return layout.Flex{
		Axis: layout.Vertical,
		// Spacing: layout.SpaceEvenly,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return ui.q100_LabelValue(gtx, name[0], value[0])
		}),
		layout.Rigid(func(gtx C) D {
			return ui.q100_LabelValue(gtx, name[1], value[1])
		}),
		layout.Rigid(func(gtx C) D {
			return ui.q100_LabelValue(gtx, name[2], value[2])
		}),
		layout.Rigid(func(gtx C) D {
			return ui.q100_LabelValue(gtx, name[3], value[3])
		}),
	)
}

// returns a column with 2 buttons
func (ui *UI) q100_Column2Buttons(gtx C) D {
	const btnWidth = 70
	const btnHeight = 50
	inset := layout.Inset{
		Top:    2,
		Bottom: 2,
		Left:   4,
		Right:  4,
	}
	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceEvenly,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(btnWidth)
				gtx.Constraints.Min.Y = gtx.Dp(btnHeight)
				return ui.q100_Button(gtx, &ui.tune, "TUNE", rxData.CurIsTuned, q100color.buttonGreen)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(btnWidth)
				gtx.Constraints.Min.Y = gtx.Dp(btnHeight)
				return ui.q100_Button(gtx, &ui.stream, "Stream", rxData.CurIsStreaming, q100color.buttonRed)
			})
		}),
	)
}

// Returns a 3x4 matrix of status + 1 column with 2 buttons
func (ui *UI) q100_3x4statusMatrixPlus2buttons(gtx C) D {
	names1 := [4]string{"Frequency", "Symbol Rate", "Mode", "Constellation"}
	values1 := [4]string{lmData.Frequency, lmData.SymbolRate, lmData.Mode, lmData.Constellation}

	names2 := [4]string{"FEC", "Codecs", "dB MER", "dB Margin"}
	values2 := [4]string{lmData.Fec, lmData.VideoCodec + " " + lmData.AudioCodec, lmData.DbMer, lmData.DbMargin}

	names3 := [4]string{"dBm Power", "Null Ratio %", "Video PID", "Audio PID"}
	values3 := [4]string{lmData.DbmPower, lmData.NullRatio, lmData.PidPair1, lmData.PidPair2}

	return layout.Flex{
		Axis: layout.Horizontal,
		// Spacing: layout.SpaceEvenly,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return ui.q100_Column4Rows(gtx, names1, values1)
		}),
		layout.Rigid(func(gtx C) D {
			return ui.q100_Column4Rows(gtx, names2, values2)
		}),
		layout.Rigid(func(gtx C) D {
			return ui.q100_Column4Rows(gtx, names3, values3)
		}),
		layout.Rigid(func(gtx C) D {
			return ui.q100_Column2Buttons(gtx)
		}),
	)
}

// layoutFlexes returns the entire display
func (ui *UI) layoutFlexes(gtx C) D {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			return layout.Flex{
				Axis: layout.Vertical,
				// Spacing:   layout.SpaceEnd,
				// Alignment: layout.Alignment(layout.N),
			}.Layout(gtx,
				layout.Rigid(ui.q100_TopStatusRow),
				layout.Rigid(ui.q100_SpectrumDisplay),
				layout.Rigid(ui.q100_MainTuningRow),
				layout.Rigid(ui.q100_3x4statusMatrixPlus2buttons),
			)
		}),
	)
}
