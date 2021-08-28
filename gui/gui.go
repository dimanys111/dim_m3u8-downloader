package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func New(download func(string, string, chan string, chan string)) fyne.App {
	a := app.New()
	w := a.NewWindow("m3u8 downloader")
	w.Resize(fyne.Size{Width: 500, Height: 500})

	urlInput := widget.NewEntry()
	fnInput := widget.NewEntry()

	logInfo := widget.NewTextGrid()
	logInfo.Resize(fyne.Size{Height: 500})

	inChan := make(chan string)
	outChan := make(chan string)

	downloaderState := make(chan string, 4)
	go func() {
		for {
			msg := <-downloaderState
			logInfo.SetRow(len(logInfo.Rows)+1, textToGridRow(msg))
		}
	}()

	btn := widget.NewButton("Download", nil)
	btn_st := widget.NewButton("Stop", nil)
	btn_st.Disable()
	btn.OnTapped = func() {
		btn.Disable()
		btn_st.Enable()
		url := urlInput.Text
		fn := fnInput.Text
		downloaderState <- fmt.Sprintf("Start download %s", url)
		go download(url, fn, inChan, outChan)

		go func() {
			for {
				msg := <-outChan
				downloaderState <- fmt.Sprintf("Download " + msg)
				btn.Enable()
				btn_st.Disable()
			}
		}()
	}

	btn_st.OnTapped = func() {
		inChan <- "stop"
		btn_st.Disable()
	}

	w.SetContent(
		container.New(layout.NewGridLayout(1),
			container.New(layout.NewFormLayout(),
				widget.NewLabel("URL"),
				urlInput,
			),
			container.New(layout.NewFormLayout(),
				widget.NewLabel("FN"),
				fnInput,
			),
			container.New(layout.NewCenterLayout(), btn),
			container.New(layout.NewCenterLayout(), btn_st),
			container.NewScroll(logInfo),
		),
	)
	w.Show()

	return a
}

func textToGridRow(msg string) widget.TextGridRow {
	var cells []widget.TextGridCell
	for _, r := range []rune(msg) {
		cells = append(cells, widget.TextGridCell{Rune: r})
	}
	return widget.TextGridRow{
		Cells: cells,
	}
}
