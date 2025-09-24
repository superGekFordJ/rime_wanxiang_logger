package ui

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

var (
	// Basic styles
	Title      = color.New(color.Bold, color.FgCyan).SprintFunc()
	Subtitle   = color.New(color.Bold).SprintFunc()
	SuccessTxt = color.New(color.FgGreen).SprintFunc()
	ErrorTxt   = color.New(color.FgRed).SprintFunc()
	WarnTxt    = color.New(color.FgYellow).SprintFunc()
	InfoTxt    = color.New(color.FgBlue).SprintFunc()
)

// Section prints a prominent section header.
func Section(title string) {
	sep := strings.Repeat("─", len(title)+2)
	fmt.Printf("\n%s\n%s %s\n%s\n", color.HiCyanString(sep), color.HiCyanString("▶"), Title(title), color.HiCyanString(sep))
}

// Subsection prints a smaller subsection header.
func Subsection(title string) {
	fmt.Printf("\n%s %s\n", color.HiCyanString("→"), Subtitle(title))
}

// Simple status helpers
func Successf(format string, a ...any) {
	fmt.Printf("%s %s\n", SuccessTxt("✓"), fmt.Sprintf(format, a...))
}
func Infof(format string, a ...any) { fmt.Printf("%s %s\n", InfoTxt("i"), fmt.Sprintf(format, a...)) }
func Warnf(format string, a ...any) { fmt.Printf("%s %s\n", WarnTxt("!"), fmt.Sprintf(format, a...)) }
func Errorf(format string, a ...any) {
	fmt.Printf("%s %s\n", ErrorTxt("✗"), fmt.Sprintf(format, a...))
}

// Badge returns a colored status string like ✅/❌/❓ with optional text.
func Badge(ok bool, unknown bool, text string) string {
	switch {
	case ok:
		if text == "" {
			text = "OK"
		}
		return SuccessTxt("✅ " + text)
	case unknown:
		if text == "" {
			text = "UNKNOWN"
		}
		return WarnTxt("❓ " + text)
	default:
		if text == "" {
			text = "ERROR"
		}
		return ErrorTxt("❌ " + text)
	}
}

// PrintKV renders a two-column key/value list aligned nicely.
func PrintKV(pairs [][2]string) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, kv := range pairs {
		fmt.Fprintf(tw, "%s:\t%s\n", Subtitle(kv[0]), kv[1])
	}
	_ = tw.Flush()
}

// PrintTable renders a generic table using tabwriter.
func PrintTable(headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	// header
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, Subtitle(h))
	}
	fmt.Fprint(tw, "\n")
	// underline (approx)
	for i := range headers {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, strings.Repeat("-", 8))
	}
	fmt.Fprint(tw, "\n")
	// rows
	for _, row := range rows {
		for i, col := range row {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, col)
		}
		fmt.Fprint(tw, "\n")
	}
	_ = tw.Flush()
}
