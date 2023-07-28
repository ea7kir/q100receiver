/*
 *  Q-100 Receiver
 *  Copyright (c) 2023 Michael Naylor EA7KIR (https://michaelnaylor.es)
 */

/*********************************************************************************

[ [ button ]  [ label_____________________________________________ ]  [ button ] ]

[ [ ------------------------------- spectrum --------------------------------- ] ]

[    [ button label button ]  [ button label button ]  [ button label button ]   ]

[ [ label__  label__ ]   [ label__  label__ ]   [ label__  label__ ]  [ button ] ]
[ [ label__  label__ ]   [ label__  label__ ]   [ label__  label__ ]             ]
[ [ label__  label__ ]   [ label__  label__ ]   [ label__  label__ ]             ]
[ [ label__  label__ ]   [ label__  label__ ]   [ label__  label__ ]  [ button ] ]

*********************************************************************************/

package main

import (
	"context"
	"image"
	"image/color"
	"os"
	"os/signal"
	"q100receiver/lmClient"
	"q100receiver/logger"
	"q100receiver/rxControl"
	"q100receiver/spectrumClient"

	"github.com/ajstarks/giocanvas"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/image/colornames"
)

// application directory for the configuration data
const appFolder = "/home/pi/Q100/q100receiver/"

// configuration data
var (
	spectrumConfig = spectrumClient.SpConfig{
		Url: "wss://eshail.batc.org.uk/wb/fft/fft_ea7kirsatcontroller:443/",
	}
	lmConfig = lmClient.LmConfig{
		Folder:      appFolder + "_longmynd/",
		Binary:      appFolder + "_longmynd/longmynd",
		Offset:      float64(9750000),
		StatusFifo:  appFolder + "_longmynd/longmynd_main_status",
		StartScript: appFolder + "_scripts/longmyndrun",
		StopScript:  appFolder + "_scripts/longmyndkill",
	}
	fpConfig = lmClient.FpConfig{
		Binary:      "/usr/bin/ffplay",
		TsFifo:      appFolder + "_longmynd/longmynd_main_ts",
		Volume:      "70",
		StartScript: appFolder + "_scripts/ffplayrun",
		StopScript:  appFolder + "_scripts/ffplaykill",
	}
	tuConfig = rxControl.TuConfig{
		Band:                 "Narrow",
		WideSymbolrate:       "1000",
		NarrowSymbolrate:     "333",
		VeryNarrowSymbolRate: "125",
		WideFrequency:        "10494.75 / 09",
		NarrowFrequency:      "10499.25 / 27",
		VeryNarrowFrequency:  "10496.00 / 14",
	}
)

// local data
var (
	spData    spectrumClient.SpData
	spChannel = make(chan spectrumClient.SpData, 5)
	lmData    lmClient.LongmyndData
	lmChannel = make(chan lmClient.LongmyndData, 5)
)

// main - with some help from Chris Waldon who got me started
func main() {
	logger.Open("/home/pi/Q100/receiver.log")
	defer logger.Close()

	os.Setenv("DISPLAY", ":0") // required for X11

	spectrumClient.Intitialize(spectrumConfig, spChannel)

	rxControl.Intitialize(tuConfig)

	lmClient.Intitialize(lmConfig, fpConfig, lmChannel)

	go func() {
		w := app.NewWindow(app.Fullscreen.Option())
		app.Size(800, 480) // I don't know if this is help in any way
		if err := loop(w); err != nil {
			logger.Fatal.Fatalf(": ", err)
		}

		rxControl.Stop()
		lmClient.Stop()
		spectrumClient.Stop()

		os.Exit(0)
	}()

	app.Main()
}

func loop(w *app.Window) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ui := UI{
		th: material.NewTheme(gofont.Collection()),
	}
	var ops op.Ops
	// Capture the context done channel in a variable so that we can nil it
	// out after it closes and prevent its select case from firing again.
	done := ctx.Done()

	for {
		select {
		case <-done:
			// When the context cancels, assign the done channel to nil to
			// prevent it from firing over and over.
			done = nil
			// Log something to make it obvious this happened.
			// logger.Info("context cancelled")
			// Initiate window shutdown.
			rxControl.Stop()      // TODO: does nothing yet
			lmClient.Stop()       // TODO: does nothing yet
			spectrumClient.Stop() // TODO: does nothing yet - bombs with Control=C
			w.Perform(system.ActionClose)
		case lmData = <-lmChannel:
			w.Invalidate()
		case spData = <-spChannel:
			w.Invalidate()
		case event := <-w.Events():
			switch event := event.(type) {
			case system.DestroyEvent:
				return event.Err
			case system.FrameEvent:
				if ui.about.Clicked() {
					showAboutBox()
				}
				if ui.shutdown.Clicked() {
					w.Perform(system.ActionClose)
				}
				if ui.decBand.Clicked() {
					rxControl.DecBandSelector(&rxControl.Band)
				}
				if ui.incBand.Clicked() {
					rxControl.IncBandSelector(&rxControl.Band)
				}
				if ui.decSymbolRate.Clicked() {
					rxControl.DecSelector(&rxControl.SymbolRate)
				}
				if ui.incSymbolRate.Clicked() {
					rxControl.IncSelector(&rxControl.SymbolRate)
				}
				if ui.decFrequency.Clicked() {
					rxControl.DecSelector(&rxControl.Frequency)
				}
				if ui.incFrequency.Clicked() {
					rxControl.IncSelector(&rxControl.Frequency)
				}
				if ui.tune.Clicked() {
					rxControl.Tune()
				}
				if ui.mute.Clicked() {
					rxControl.Mute()
				}

				gtx := layout.NewContext(&ops, event)
				// set the screen background to dark grey
				paint.Fill(gtx.Ops, q100color.screenGrey)
				ui.layoutFlexes(gtx)
				event.Frame(gtx.Ops)
			}
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
	gfxMarker:    color.NRGBA{R: 10, G: 10, B: 10, A: 255},
	gfxGraticule: color.NRGBA(colornames.Darkgray),
	gfxLabel:     color.NRGBA{R: 32, G: 32, B: 32, A: 255}, // DarkGrey is too light
}

// define all buttons
type UI struct {
	about, shutdown              widget.Clickable
	decBand, incBand             widget.Clickable
	decSymbolRate, incSymbolRate widget.Clickable
	decFrequency, incFrequency   widget.Clickable
	tune, mute                   widget.Clickable
	th                           *material.Theme
}

// makes the code more readable
type (
	C = layout.Context
	D = layout.Dimensions
)

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
	const btnWidth = 30
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
				return ui.q100_Button(gtx, &ui.about, "Q-100 Receiver", false, q100color.buttonGrey)
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

// Returns a single Selector as [ button label button ]
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
		// Alignment: layout.Middle,
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
			return ui.q100_Selector(gtx, &ui.decBand, &ui.incBand, rxControl.Band.Value, btnWidth, 100)
		}),
		layout.Rigid(func(gtx C) D {
			return ui.q100_Selector(gtx, &ui.decSymbolRate, &ui.incSymbolRate, rxControl.SymbolRate.Value, btnWidth, 50)
		}),
		layout.Rigid(func(gtx C) D {
			return ui.q100_Selector(gtx, &ui.decFrequency, &ui.incFrequency, rxControl.Frequency.Value, btnWidth, 100)
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
				}
				// fmt("  Canvas: %#v\n", canvas.Context.Constraints)

				canvas.Background(q100color.gfxBgd)
				// tuning marker
				canvas.Rect(spData.MarkerCentre, 50, spData.MarkerWidth, 100, q100color.gfxMarker)
				// polygon
				canvas.Polygon(spectrumClient.Xp, spData.Yp, q100color.gfxGreen)
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
				return ui.q100_Button(gtx, &ui.tune, "TUNE", rxControl.IsTuned, q100color.buttonGreen)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(btnWidth)
				gtx.Constraints.Min.Y = gtx.Dp(btnHeight)
				return ui.q100_Button(gtx, &ui.mute, "MUTE", rxControl.IsMuted, q100color.buttonRed)
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
	names3 := [4]string{"dBm Power", "Null Ratio", "Provider", "Service"}
	values3 := [4]string{lmData.DbmPower, lmData.NullRatio, lmData.Provider, lmData.Service}

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
