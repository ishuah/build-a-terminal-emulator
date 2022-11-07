package main

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/creack/pty"
)

// MaxBufferSize sets the size limit
// for our command output buffer.
const MaxBufferSize = 16

func main() {
	a := app.New()
	w := a.NewWindow("germ")

	ui := widget.NewTextGrid() // Create a new TextGrid

	os.Setenv("TERM", "dumb")
	c := exec.Command("/bin/bash")
	p, err := pty.Start(c)

	if err != nil {
		fyne.LogError("Failed to open pty", err)
		os.Exit(1)
	}

	defer c.Process.Kill()

	// Callback function that handles special keypresses
	onTypedKey := func(e *fyne.KeyEvent) {
		if e.Name == fyne.KeyEnter || e.Name == fyne.KeyReturn {
			_, _ = p.Write([]byte{'\r'})
		}
	}

	// Callback function that handles character keypresses
	onTypedRune := func(r rune) {
		_, _ = p.WriteString(string(r))
	}

	w.Canvas().SetOnTypedKey(onTypedKey)
	w.Canvas().SetOnTypedRune(onTypedRune)

	buffer := [][]rune{}
	reader := bufio.NewReader(p)

	// Goroutine that reads from pty
	go func() {
		line := []rune{}
		buffer = append(buffer, line)
		for {
			r, _, err := reader.ReadRune()

			if err != nil {
				if err == io.EOF {
					return
				}
				os.Exit(0)
			}

			line = append(line, r)
			buffer[len(buffer)-1] = line
			if r == '\n' {
				if len(buffer) > MaxBufferSize { // If the buffer is at capacity...
					buffer = buffer[1:] // ...pop the first line in the buffer
				}

				line = []rune{}
				buffer = append(buffer, line)
			}
		}
	}()

	// Goroutine that renders to UI
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			ui.SetText("")
			var lines string
			for _, line := range buffer {
				lines = lines + string(line)
			}
			ui.SetText(string(lines))
		}
	}()

	// Create a new container with a wrapped layout
	// set the layout width to 900, height to 325
	w.SetContent(
		fyne.NewContainerWithLayout(
			layout.NewGridWrapLayout(fyne.NewSize(900, 325)),
			ui,
		),
	)
	w.ShowAndRun()
}
