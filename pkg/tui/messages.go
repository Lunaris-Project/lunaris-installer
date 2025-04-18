package tui

import "time"

// InstallProgressMsg represents a message for installation progress updates
type InstallProgressMsg struct {
	Progress               int
	Total                  int
	CurrentStep            string
	Error                  error
	Phase                  string
	IsComplete             bool
	HasConflict            bool
	Conflict               string
	IsDotfilesConfirmation bool
	IsBackupConfirmation   bool
}

// PageTransitionMsg represents a message for page transitions with animation
type PageTransitionMsg struct {
	FromPage Page
	ToPage   Page
	AnimType string
	Duration time.Duration
}

// NewInstallProgressMsg creates a new InstallProgressMsg
func NewInstallProgressMsg(progress, total int, currentStep, phase string, err error) InstallProgressMsg {
	return InstallProgressMsg{
		Progress:               progress,
		Total:                  total,
		CurrentStep:            currentStep,
		Phase:                  phase,
		Error:                  err,
		IsComplete:             false,
		HasConflict:            false,
		IsDotfilesConfirmation: false,
		IsBackupConfirmation:   false,
	}
}

// NewCompleteMsg creates a new InstallProgressMsg indicating completion
func NewCompleteMsg() InstallProgressMsg {
	return InstallProgressMsg{
		IsComplete: true,
	}
}

// NewConflictMsg creates a new InstallProgressMsg indicating a conflict
func NewConflictMsg(conflict string) InstallProgressMsg {
	return InstallProgressMsg{
		HasConflict: true,
		Conflict:    conflict,
	}
}

// NewDotfilesConfirmationMsg creates a new InstallProgressMsg for dotfiles confirmation
func NewDotfilesConfirmationMsg() InstallProgressMsg {
	return InstallProgressMsg{
		IsDotfilesConfirmation: true,
	}
}

// NewBackupConfirmationMsg creates a new InstallProgressMsg for backup confirmation
func NewBackupConfirmationMsg() InstallProgressMsg {
	return InstallProgressMsg{
		IsBackupConfirmation: true,
	}
}

// NewPageTransitionMsg creates a new PageTransitionMsg
func NewPageTransitionMsg(fromPage, toPage Page, animType string, duration time.Duration) PageTransitionMsg {
	return PageTransitionMsg{
		FromPage: fromPage,
		ToPage:   toPage,
		AnimType: animType,
		Duration: duration,
	}
}
