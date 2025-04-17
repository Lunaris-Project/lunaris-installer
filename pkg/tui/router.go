package tui

import (
	"time"

	"github.com/Lunaris-Project/lunaris-installer/pkg/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// Route represents a route in the application
type Route struct {
	Page     Page
	Title    string
	Renderer func() string
	Updater  func(tea.KeyMsg) (tea.Model, tea.Cmd)
}

// Router manages the application routes
type Router struct {
	routes       map[Page]Route
	history      []Page
	currentPage  Page
	transitions  map[Page]map[Page]func() tea.Cmd
	errorHandler func(error) tea.Cmd
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		routes:      make(map[Page]Route),
		history:     []Page{},
		currentPage: WelcomePage,
		transitions: make(map[Page]map[Page]func() tea.Cmd),
	}
}

// RegisterRoute registers a route
func (r *Router) RegisterRoute(route Route) {
	r.routes[route.Page] = route
}

// RegisterTransition registers a transition between pages
func (r *Router) RegisterTransition(from, to Page, handler func() tea.Cmd) {
	if _, ok := r.transitions[from]; !ok {
		r.transitions[from] = make(map[Page]func() tea.Cmd)
	}
	r.transitions[from][to] = handler
}

// SetErrorHandler sets the error handler
func (r *Router) SetErrorHandler(handler func(error) tea.Cmd) {
	r.errorHandler = handler
}

// Navigate navigates to a page with animation
func (r *Router) Navigate(to Page, m Model) (tea.Model, tea.Cmd) {
	// Check if the page exists
	if _, ok := r.routes[to]; !ok {
		if r.errorHandler != nil {
			return m, r.errorHandler(ErrPageNotFound{Page: to})
		}
		return m, nil
	}

	// Create a transition animation
	from := r.currentPage

	// Add current page to history
	r.history = append(r.history, r.currentPage)

	// Update current page
	r.currentPage = to

	// Create a command to start the animation
	animCmd := func() tea.Msg {
		return NewPageTransitionMsg(from, to, ui.SlideLeft, 300*time.Millisecond)
	}

	// Check if there's a transition handler
	var transitionCmd tea.Cmd
	if fromHandlers, ok := r.transitions[from]; ok {
		if handler, ok := fromHandlers[to]; ok {
			// Execute transition handler
			transitionCmd = handler()
		}
	}

	// Return both commands
	if transitionCmd != nil {
		return m, tea.Batch(animCmd, transitionCmd)
	}

	// Default transition with animation only
	return m, animCmd
}

// Back navigates back to the previous page with animation
func (r *Router) Back(m Model) (tea.Model, tea.Cmd) {
	if len(r.history) == 0 {
		return m, nil
	}

	// Get the previous page
	previousPage := r.history[len(r.history)-1]
	r.history = r.history[:len(r.history)-1]

	// Create a transition animation
	from := r.currentPage

	// Update current page
	r.currentPage = previousPage

	// Create a command to start the animation
	animCmd := func() tea.Msg {
		return NewPageTransitionMsg(from, previousPage, ui.SlideRight, 300*time.Millisecond)
	}

	return m, animCmd
}

// CurrentPage returns the current page
func (r *Router) CurrentPage() Page {
	return r.currentPage
}

// GetRoute returns the route for a page
func (r *Router) GetRoute(page Page) (Route, bool) {
	route, ok := r.routes[page]
	return route, ok
}

// ErrPageNotFound is returned when a page is not found
type ErrPageNotFound struct {
	Page Page
}

// Error returns the error message
func (e ErrPageNotFound) Error() string {
	return "page not found: " + string(e.Page)
}
