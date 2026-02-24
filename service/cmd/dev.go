package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	spinnerInterval = 120 * time.Millisecond
	durationRound   = 100 * time.Millisecond
	minRowColumns   = 2
	tablePadding    = 2
)

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiGreen  = "\x1b[32m"
	ansiRed    = "\x1b[31m"
	ansiBlue   = "\x1b[34m"
	ansiCyan   = "\x1b[36m"
	ansiGray   = "\x1b[90m"
	ansiYellow = "\x1b[33m"
)

func init() {
	devCmd := &cobra.Command{
		Use:   "dev",
		Short: "Run local dev stack",
	}

	devCmd.AddCommand(newDevUpCmd(), newDevDownCmd(), newDevStatusCmd())
	rootCmd.AddCommand(devCmd)
}

func style(code string, bold bool, text string) string {
	prefix := code
	if bold {
		prefix = ansiBold + code
	}
	return prefix + text + ansiReset
}

func okText(text string) string {
	return style(ansiGreen, true, text)
}

func errorText(text string) string {
	return style(ansiRed, true, text)
}

func hintText(text string) string {
	return style(ansiGray, false, text)
}

func sectionText(text string) string {
	return style(ansiCyan, true, text)
}

func existingText(text string) string {
	return style(ansiGray, false, text)
}

func createdText(text string) string {
	return style(ansiGreen, true, text)
}

func warningText(text string) string {
	return style(ansiYellow, true, text)
}

func formatRows(rows [][]string) string {
	maxLabel := 0
	maxValue := 0
	for _, row := range rows {
		if len(row) < minRowColumns {
			continue
		}
		labelLen := visibleLen(row[0])
		valueLen := visibleLen(row[1])
		if labelLen > maxLabel {
			maxLabel = labelLen
		}
		if valueLen > maxValue {
			maxValue = valueLen
		}
	}

	if maxLabel == 0 && maxValue == 0 {
		return ""
	}

	labelWidth := maxLabel + tablePadding
	valueWidth := maxValue + tablePadding
	top := "┌" + strings.Repeat("─", labelWidth) + "┬" + strings.Repeat("─", valueWidth) + "┐\n"
	mid := "├" + strings.Repeat("─", labelWidth) + "┼" + strings.Repeat("─", valueWidth) + "┤\n"
	bot := "└" + strings.Repeat("─", labelWidth) + "┴" + strings.Repeat("─", valueWidth) + "┘\n"

	var out strings.Builder
	out.WriteString(top)
	for _, row := range rows {
		if len(row) < minRowColumns {
			continue
		}
		label := padRight(row[0], labelWidth-1)
		value := padRight(row[1], valueWidth-1)
		fmt.Fprintf(&out, "│ %s│ %s│\n", label, value)
		out.WriteString(mid)
	}
	table := out.String()
	return strings.TrimSuffix(table, mid) + bot
}

func padRight(text string, width int) string {
	padding := width - visibleLen(text)
	if padding <= 0 {
		return text
	}
	return text + strings.Repeat(" ", padding)
}

func visibleLen(text string) int {
	if text == "" {
		return 0
	}
	n := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\x1b' && i+1 < len(text) && text[i+1] == '[' {
			i += 2
			for i < len(text) && text[i] != 'm' {
				i++
			}
			continue
		}
		n++
	}
	return n
}

func runSpinnerStep(out io.Writer, label string, fn func() error) error {
	ticker := time.NewTicker(spinnerInterval)
	defer ticker.Stop()

	start := time.Now()
	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	frames := []string{"|", "/", "-", "\\"}
	frameIdx := 0
	for {
		select {
		case err := <-done:
			duration := time.Since(start).Round(durationRound)
			durationSuffix := hintText("(" + duration.String() + ")")
			if err != nil {
				fmt.Fprintf(out, "\r%s %s %s: %v\n", errorText("[error]"), label, durationSuffix, err)
				return err
			}
			fmt.Fprintf(out, "\r%s %s %s\n", okText("[ok]"), label, durationSuffix)
			return nil
		case <-ticker.C:
			frame := frames[frameIdx%len(frames)]
			frameIdx++
			fmt.Fprintf(out, "\r%s %s", style(ansiBlue, true, frame), label)
		}
	}
}
