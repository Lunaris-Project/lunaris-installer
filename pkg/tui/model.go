package tui

import (
	"github.com/Lunaris-Project/lunaris-installer/pkg/aur"
	"github.com/Lunaris-Project/lunaris-installer/pkg/config"
	"github.com/Lunaris-Project/lunaris-installer/pkg/tui/messages"
	"github.com/Lunaris-Project/lunaris-installer/pkg/tui/ui"
	"github.com/charmbracelet/bubbles/help"
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

// Import KeyMap from keymap.go

// Model represents the state of the application
type Model struct {
	// Core components
	keyMap          KeyMap
	help            help.Model
	spinner         spinner.Model
	width           int
	height          int
	page            Page
	router          *Router
	messageQueue    *messages.Queue
	messageRenderer *messages.Renderer

	// Animation
	animation   ui.AnimationState
	animating   bool
	prevContent string
	nextContent string

	// AUR helper
	aurHelperOptions   []string
	aurHelperIndex     int
	aurHelper          *aur.Helper
	aurHelperInstalled bool // Track if the AUR helper is installed

	// Package selection
	categories       []config.PackageCategory
	categoryIndex    int
	optionIndex      int
	selectedOptions  map[string][]string
	selectedCategory int

	// Search
	searchQuery     string
	searchFocused   bool
	filteredOptions []string

	// Installation state
	installProgress   int
	installTotal      int
	installCurrent    string
	installError      string
	installComplete   bool
	packagesToInstall []string
	totalSteps        int
	currentStep       string
	installPhase      string
	errorMessage      string

	// Task progress
	tasks            []ui.TaskProgress
	indeterminatePos int

	// UI state
	showHelp         bool
	passwordInput    string
	awaitingPassword bool
	passwordVisible  bool

	// Notifications
	notifications     []ui.Notification
	showNotifications bool

	// Conflict resolution
	hasConflict        bool
	conflictMessage    string
	conflictChoice     bool
	conflictOption     int // 0=Skip, 1=Replace, 2=All, 3=Cancel
	conflictPackage    string
	skippedPackages    map[string]bool // Track packages to skip
	replaceAllPackages bool            // Track if we should replace all packages

	// Installation phases
	installationPhase    string   // Current installation phase: "packages" or "post-installation"
	phaseMessageShown    bool     // Track if we've shown the phase transition message
	repoCloned           bool     // Track if we've cloned the repository
	configDirIndex       int      // Track which config directory we're currently processing
	dotfilesConfirmation bool     // Track if the user wants to install dotfiles
	backupConfirmation   bool     // Track if the user wants to backup existing config
	systemMessages       []string // Store system messages for display (legacy, will be replaced by messageQueue)
}

// NewModel creates a new model
func NewModel() Model {
	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ui.PrimaryColor).Bold(true)

	// Initialize router
	router := NewRouter()

	// Initialize message queue and renderer
	messageQueue := messages.NewQueue(100)          // Store up to 100 messages
	messageRenderer := messages.NewRenderer(80, 15) // Default width and height

	// Set message styles
	messageRenderer.SetStyle(messages.InfoMessage, lipgloss.NewStyle().Foreground(ui.TextColor))
	messageRenderer.SetStyle(messages.SuccessMessage, lipgloss.NewStyle().Foreground(ui.SuccessColor).Bold(true))
	messageRenderer.SetStyle(messages.WarningMessage, lipgloss.NewStyle().Foreground(ui.WarningColor).Bold(true))
	messageRenderer.SetStyle(messages.ErrorMessage, lipgloss.NewStyle().Foreground(ui.ErrorColor).Bold(true))
	messageRenderer.SetStyle(messages.DebugMessage, lipgloss.NewStyle().Foreground(ui.DimmedColor))

	// Create model
	m := Model{
		keyMap:               DefaultKeyMap(),
		help:                 help.New(),
		spinner:              s,
		page:                 WelcomePage,
		router:               router,
		messageQueue:         messageQueue,
		messageRenderer:      messageRenderer,
		animation:            ui.AnimationState{},
		animating:            false,
		prevContent:          "",
		nextContent:          "",
		aurHelperOptions:     config.AURHelpers,
		aurHelperIndex:       0,
		aurHelperInstalled:   false,
		categories:           config.PackageCategories,
		categoryIndex:        0,
		optionIndex:          -1,
		selectedOptions:      make(map[string][]string),
		selectedCategory:     0,
		searchQuery:          "",
		searchFocused:        false,
		filteredOptions:      []string{},
		installProgress:      0,
		installTotal:         0,
		installCurrent:       "",
		installError:         "",
		installComplete:      false,
		tasks:                make([]ui.TaskProgress, 0),
		indeterminatePos:     0,
		showHelp:             false,
		passwordInput:        "",
		awaitingPassword:     false,
		passwordVisible:      true,
		notifications:        make([]ui.Notification, 0),
		showNotifications:    true,
		hasConflict:          false,
		conflictMessage:      "",
		conflictChoice:       true,
		conflictOption:       0,
		conflictPackage:      "",
		skippedPackages:      make(map[string]bool),
		replaceAllPackages:   false,
		installationPhase:    "",
		phaseMessageShown:    false,
		repoCloned:           false,
		configDirIndex:       0,
		dotfilesConfirmation: false,
		backupConfirmation:   false,
		systemMessages:       make([]string, 0),
		packagesToInstall:    make([]string, 0),
	}

	// Register routes
	router.RegisterRoute(Route{
		Page:     WelcomePage,
		Title:    "Welcome",
		Renderer: m.renderWelcomePage,
		Updater:  m.updateWelcomePage,
	})

	router.RegisterRoute(Route{
		Page:     AURHelperPage,
		Title:    "AUR Helper",
		Renderer: m.renderAURHelperPage,
		Updater:  m.updateAURHelperPage,
	})

	router.RegisterRoute(Route{
		Page:     PackageCategoriesPage,
		Title:    "Package Categories",
		Renderer: m.renderPackageCategoriesPage,
		Updater:  m.updatePackageCategoriesPage,
	})

	router.RegisterRoute(Route{
		Page:     InstallationPage,
		Title:    "Installation",
		Renderer: m.renderInstallationPage,
		Updater:  m.updateInstallationPage,
	})

	router.RegisterRoute(Route{
		Page:     CompletePage,
		Title:    "Complete",
		Renderer: m.renderCompletePage,
		Updater:  m.updateCompletePage,
	})

	// Register transitions
	router.RegisterTransition(WelcomePage, AURHelperPage, func() tea.Cmd {
		return m.AddInfoNotification("Welcome", "Please select your preferred AUR helper")
	})

	router.RegisterTransition(AURHelperPage, PackageCategoriesPage, func() tea.Cmd {
		return m.AddInfoNotification("AUR Helper Selected", "Now select the packages you want to install")
	})

	router.RegisterTransition(PackageCategoriesPage, InstallationPage, func() tea.Cmd {
		return tea.Batch(
			m.AddSuccessNotification("Installation Started", "Installing selected packages"),
			m.startInstallation(),
		)
	})

	router.RegisterTransition(InstallationPage, CompletePage, func() tea.Cmd {
		return m.AddSuccessNotification("Installation Complete", "All packages have been installed successfully")
	})

	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.tickIndeterminateProgress(),
	)
}
