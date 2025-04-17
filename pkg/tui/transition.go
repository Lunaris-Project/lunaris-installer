package tui

import (
	"time"

	"github.com/Lunaris-Project/lunaris-installer/pkg/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// handlePageTransition handles page transitions with animation
func (m Model) handlePageTransition(msg PageTransitionMsg) (tea.Model, tea.Cmd) {
	// Skip animation if width is not set yet (window size not received)
	if m.width == 0 {
		return m, nil
	}

	// Start the animation
	m.animation = ui.NewAnimation(msg.AnimType, msg.Duration)
	m.animating = true

	// Render the previous page content
	prevRoute, ok := m.router.GetRoute(msg.FromPage)
	if ok {
		m.prevContent = prevRoute.Renderer()
	}

	// Render the next page content
	nextRoute, ok := m.router.GetRoute(msg.ToPage)
	if ok {
		m.nextContent = nextRoute.Renderer()
	}

	// Create a command to update the animation
	return m, m.updateAnimation()
}

// updateAnimation updates the animation state
func (m Model) updateAnimation() tea.Cmd {
	return tea.Tick(time.Millisecond*16, func(t time.Time) tea.Msg {
		return animationTickMsg{}
	})
}

// animationTickMsg is a message for animation ticks
type animationTickMsg struct{}

// handleAnimationTick handles animation ticks
func (m Model) handleAnimationTick() (tea.Model, tea.Cmd) {
	// Update the animation state
	isActive := m.animation.Update()

	// If the animation is complete, stop animating
	if !isActive {
		m.animating = false
		m.prevContent = ""
		return m, nil
	}

	// Continue the animation
	return m, m.updateAnimation()
}
