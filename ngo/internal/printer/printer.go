package printer

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	success  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	warn     = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	muted    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func Success(msg string) { fmt.Println(success.Render("✓ " + msg)) }
func Warn(msg string)    { fmt.Println(warn.Render("⚠ " + msg)) }
func Error(msg string)   { fmt.Fprintln(os.Stderr, errStyle.Render("✗ "+msg)) }
func Info(msg string)    { fmt.Println(muted.Render("  " + msg)) }
