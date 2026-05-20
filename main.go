package main

import (
	"fmt"
	"os"

	"fast-go/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	fmt.Println("fast-go — Local Speed Test")

	p := tea.NewProgram(
		tui.InitialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
