package ui

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Animation types
const (
	FadeIn     = "fade-in"
	FadeOut    = "fade-out"
	SlideLeft  = "slide-left"
	SlideRight = "slide-right"
	SlideUp    = "slide-up"
	SlideDown  = "slide-down"
)

// AnimationState represents the state of an animation
type AnimationState struct {
	Type      string
	Progress  float64
	StartTime time.Time
	Duration  time.Duration
	IsActive  bool
}

// NewAnimation creates a new animation
func NewAnimation(animType string, duration time.Duration) AnimationState {
	return AnimationState{
		Type:      animType,
		Progress:  0.0,
		StartTime: time.Now(),
		Duration:  duration,
		IsActive:  true,
	}
}

// Update updates the animation state
func (a *AnimationState) Update() bool {
	if !a.IsActive {
		return false
	}

	elapsed := time.Since(a.StartTime)
	a.Progress = float64(elapsed) / float64(a.Duration)

	if a.Progress >= 1.0 {
		a.Progress = 1.0
		a.IsActive = false
	}

	return a.IsActive
}

// Reset resets the animation
func (a *AnimationState) Reset() {
	a.Progress = 0.0
	a.StartTime = time.Now()
	a.IsActive = true
}

// AnimateContent applies animation to content
func AnimateContent(content string, animation AnimationState, width, height int) string {
	if !animation.IsActive {
		return content
	}

	switch animation.Type {
	case FadeIn:
		return fadeIn(content, animation.Progress)
	case FadeOut:
		return fadeOut(content, animation.Progress)
	case SlideLeft:
		return slideLeft(content, animation.Progress, width)
	case SlideRight:
		return slideRight(content, animation.Progress, width)
	case SlideUp:
		return slideUp(content, animation.Progress, height)
	case SlideDown:
		return slideDown(content, animation.Progress, height)
	default:
		return content
	}
}

// fadeIn implements a fade-in animation
func fadeIn(content string, progress float64) string {
	// Calculate opacity based on progress
	opacity := progress

	// Apply opacity to content
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#000000")).
		Faint(!((opacity) > 0.5))

	return style.Render(content)
}

// fadeOut implements a fade-out animation
func fadeOut(content string, progress float64) string {
	// Calculate opacity based on progress
	opacity := 1.0 - progress

	// Apply opacity to content
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#000000")).
		Faint(!((opacity) > 0.5))

	return style.Render(content)
}

// slideLeft implements a slide-left animation
func slideLeft(content string, progress float64, width int) string {
	// Calculate offset based on progress
	offset := int(float64(width) * (1.0 - progress))

	// Apply offset to content
	style := lipgloss.NewStyle().
		PaddingLeft(offset)

	return style.Render(content)
}

// slideRight implements a slide-right animation
func slideRight(content string, progress float64, width int) string {
	// Calculate offset based on progress
	offset := int(float64(width) * progress)

	// Apply offset to content
	style := lipgloss.NewStyle().
		PaddingLeft(offset)

	return style.Render(content)
}

// slideUp implements a slide-up animation
func slideUp(content string, progress float64, height int) string {
	// Calculate offset based on progress
	offset := int(float64(height) * (1.0 - progress))

	// Apply offset to content
	style := lipgloss.NewStyle().
		PaddingTop(offset)

	return style.Render(content)
}

// slideDown implements a slide-down animation
func slideDown(content string, progress float64, height int) string {
	// Calculate offset based on progress
	offset := int(float64(height) * progress)

	// Apply offset to content
	style := lipgloss.NewStyle().
		PaddingTop(offset)

	return style.Render(content)
}
