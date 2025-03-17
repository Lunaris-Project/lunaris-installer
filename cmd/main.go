package main

import (
	"fmt"
	"os"

	"github.com/Lunaris-Project/lunaris-installer/pkg/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Create a new model
	m := tui.NewModel()

	// Initialize the program
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
