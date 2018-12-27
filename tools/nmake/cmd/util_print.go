package cmd

import (
	"github.com/fatih/color"
)

func info(s string, args ...interface{}) {
	color.Blue(s, args...)
}

func extraInfo(s string, args ...interface{}) {
	color.Yellow(s, args...)
}
