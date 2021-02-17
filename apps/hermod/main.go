package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"nimona.io/apps/hermod/tui"
)

type Config struct {
	ReceivedFolder string `envconfig:"RECEIVED_FOLDER" default:"received_files"`
}

func main() {
	h := tui.NewHermod()

	p := tea.NewProgram(h)
	p.EnterAltScreen()

	err := p.Start()
	p.ExitAltScreen()
	if err != nil {
		fmt.Println("Oh no, it didn't work:", err)
		os.Exit(1)
	}
}
