package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/manifoldco/promptui"
)

var (
	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	red    = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	cyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	bold   = lipgloss.NewStyle().Bold(true)
	dim    = lipgloss.NewStyle().Faint(true)
)

func Success(msg string) { fmt.Println(green.Render("✓ " + msg)) }
func Error(msg string)   { fmt.Fprintln(os.Stderr, red.Render("✗ "+msg)) }
func Warn(msg string)    { fmt.Println(yellow.Render("! " + msg)) }
func Info(msg string)    { fmt.Println(cyan.Render("→ " + msg)) }
func Dim(msg string)     { fmt.Println(dim.Render("  " + msg)) }
func Bold(msg string)    { fmt.Println(bold.Render(msg)) }

func Spinner(msg string) func() {
	s := spinner.New(spinner.CharSets[14], 80*time.Millisecond)
	s.Suffix = " " + msg
	s.Start()
	return s.Stop
}

func Prompt(label, defaultVal string) (string, error) {
	p := promptui.Prompt{
		Label:   label,
		Default: defaultVal,
	}
	return p.Run()
}

func PromptPassword(label string) (string, error) {
	p := promptui.Prompt{
		Label: label,
		Mask:  '*',
	}
	return p.Run()
}

func PromptSelect(label string, items []string) (string, error) {
	p := promptui.Select{
		Label: label,
		Items: items,
	}
	_, result, err := p.Run()
	return result, err
}

func PromptConfirm(label string) (bool, error) {
	p := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}
	_, err := p.Run()
	if err == promptui.ErrAbort {
		return false, nil
	}
	return err == nil, err
}

func Table(header [2]string, rows [][2]string) {
	colWidth := len(header[0])
	for _, row := range rows {
		if len(row[0]) > colWidth {
			colWidth = len(row[0])
		}
	}
	colWidth += 4

	fmt.Printf("  %-*s%s\n", colWidth, bold.Render(header[0]), bold.Render(header[1]))
	fmt.Println("  " + dim.Render(repeatStr("─", colWidth+len(header[1])+4)))
	for _, row := range rows {
		fmt.Printf("  %-*s%s\n", colWidth, row[0], dim.Render(row[1]))
	}
}

func repeatStr(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
