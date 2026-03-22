package printer

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

var (
	miniDotFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	miniDotFPS    = time.Second / 12
	spinnerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
)

func Spin(msg string) func(error) {
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		Info(msg + "...")
		return func(err error) {
			if err != nil {
				Error(msg)
			} else {
				Success(msg)
			}
		}
	}

	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(miniDotFPS)
		defer ticker.Stop()
		frame := 0
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				styled := spinnerStyle.Render(miniDotFrames[frame%len(miniDotFrames)])
				text := muted.Render(" " + msg)
				fmt.Printf("\r%s%s", styled, text)
				frame++
			}
		}
	}()

	return func(err error) {
		close(done)
		fmt.Printf("\r%s\r", strings.Repeat(" ", len(msg)+4))
		if err != nil {
			Error(msg)
		} else {
			Success(msg)
		}
	}
}
