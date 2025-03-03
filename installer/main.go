package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nixev/hyprland-installer/ui"
)

func main() {
	// Create a new model
	m := ui.NewModel()

	// Initialize the program
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
