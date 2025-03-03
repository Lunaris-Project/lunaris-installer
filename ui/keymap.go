package ui

import "github.com/charmbracelet/bubbles/key"

// ShortHelp returns keybinding help
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Quit}
}

// FullHelp returns the full set of keybindings
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Select, k.Back, k.Tab},
		{k.Help, k.Quit},
	}
}
