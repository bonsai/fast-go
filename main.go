package main

import (
	"flag"
	"fmt"
	"os"

	"fast-go/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	mini := flag.Bool("mini", false, "start in mini view mode")
	flag.Parse()

	p := tea.NewProgram(
		tui.InitialModel(*mini),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
