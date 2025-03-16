package tui

// InstallProgressMsg represents a message for installation progress updates
type InstallProgressMsg struct {
	Progress    int
	Total       int
	CurrentStep string
	Error       error
	Phase       string
	IsComplete  bool
	HasConflict bool
	Conflict    string
}

// NewInstallProgressMsg creates a new InstallProgressMsg
func NewInstallProgressMsg(progress, total int, currentStep, phase string, err error) InstallProgressMsg {
	return InstallProgressMsg{
		Progress:    progress,
		Total:       total,
		CurrentStep: currentStep,
		Phase:       phase,
		Error:       err,
		IsComplete:  false,
		HasConflict: false,
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
