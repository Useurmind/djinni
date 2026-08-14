package log

import (
	"github.com/fatih/color"
)

func Info(msg string) {
	cyan := color.New(color.FgCyan)
	_, _ = cyan.Println("[INFO]", msg)
}

func Success(msg string) {
	green := color.New(color.FgGreen)
	_, _ = green.Println("[SUCCESS]", msg)
}

func Error(msg string) {
	red := color.New(color.FgRed)
	_, _ = red.Println("[ERROR]", msg)
}
