package log

import (
	"github.com/fatih/color"
)

func Info(format string, a ...interface{}) {
	cyan := color.New(color.FgCyan)
	_, _ = cyan.Printf("[INFO] "+format+"\n", a...)
}

func Success(format string, a ...interface{}) {
	green := color.New(color.FgGreen)
	_, _ = green.Printf("[SUCCESS] "+format+"\n", a...)
}

func Error(format string, a ...interface{}) {
	red := color.New(color.FgRed)
	_, _ = red.Printf("[ERROR] "+format+"\n", a...)
}
