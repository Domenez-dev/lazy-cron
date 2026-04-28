package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/domenez-dev/lazy-chrony/internal/ui"
)

func main() {
	app, err := ui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "lazy-chrony: failed to load cron jobs: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),       // use alternate screen buffer (clean restore on exit)
		tea.WithMouseCellMotion(), // optional mouse support
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "lazy-chrony: runtime error: %v\n", err)
		os.Exit(1)
	}
}
