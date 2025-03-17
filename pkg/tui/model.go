package tui

import (
	"github.com/Lunaris-Project/lunaris-installer/pkg/aur"
	"github.com/Lunaris-Project/lunaris-installer/pkg/config"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Page represents a page in the installer
type Page int

const (
	WelcomePage Page = iota
	AURHelperPage
	PackageCategoriesPage
	InstallationPage
	CompletePage
)

// KeyMap defines the keybindings for the application
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Select key.Binding
	Back   key.Binding
	Quit   key.Binding
	Help   key.Binding
	Tab    key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Back, k.Quit, k.Help}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Select, k.Back, k.Tab},
		{k.Help, k.Quit},
	}
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("ctrl+c/q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
	}
}

// Model represents the state of the application
type Model struct {
	keyMap               KeyMap
	help                 help.Model
	spinner              spinner.Model
	width                int
	height               int
	page                 Page
	aurHelperOptions     []string
	aurHelperIndex       int
	aurHelper            *aur.Helper
	aurHelperInstalled   bool // Track if the AUR helper is installed
	categories           []config.PackageCategory
	categoryIndex        int
	optionIndex          int
	selectedOptions      map[string][]string
	selectedCategory     int
	installProgress      int
	installTotal         int
	installCurrent       string
	installError         string
	installComplete      bool
	showHelp             bool
	passwordInput        string
	awaitingPassword     bool
	passwordVisible      bool
	hasConflict          bool
	conflictMessage      string
	conflictChoice       bool
	conflictOption       int // 0=Yes, 1=No, 2=Remove
	conflictPackage      string
	skippedPackages      map[string]bool // Track packages to skip
	installationPhase    string          // Current installation phase: "packages" or "post-installation"
	phaseMessageShown    bool            // Track if we've shown the phase transition message
	repoCloned           bool            // Track if we've cloned the repository
	configDirIndex       int             // Track which config directory we're currently processing
	dotfilesConfirmation bool            // Track if the user wants to install dotfiles
	backupConfirmation   bool            // Track if the user wants to backup existing config
	systemMessages       []string        // Store system messages for display

	// Installation state
	packagesToInstall []string
	totalSteps        int
	currentStep       string
	installPhase      string
	errorMessage      string
}

// NewModel creates a new model
func NewModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)

	return Model{
		keyMap:               DefaultKeyMap(),
		help:                 help.New(),
		spinner:              s,
		page:                 WelcomePage,
		aurHelperOptions:     config.AURHelpers,
		aurHelperIndex:       0,
		aurHelperInstalled:   false,
		categories:           config.PackageCategories,
		categoryIndex:        0,
		optionIndex:          -1,
		selectedOptions:      make(map[string][]string),
		selectedCategory:     0,
		installProgress:      0,
		installTotal:         0,
		installCurrent:       "",
		installError:         "",
		installComplete:      false,
		showHelp:             false,
		passwordInput:        "",
		awaitingPassword:     false,
		passwordVisible:      true,
		hasConflict:          false,
		conflictMessage:      "",
		conflictChoice:       true,
		conflictOption:       0,
		conflictPackage:      "",
		skippedPackages:      make(map[string]bool),
		installationPhase:    "",
		phaseMessageShown:    false,
		repoCloned:           false,
		configDirIndex:       0,
		dotfilesConfirmation: false,
		backupConfirmation:   false,
		systemMessages:       make([]string, 0),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return spinner.Tick
}
