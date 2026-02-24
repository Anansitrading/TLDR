package claudewatch

// ClaudeTasksChangedMsg signals the monitor to refresh data
// because Claude Code tasks have been imported/updated.
type ClaudeTasksChangedMsg struct {
	TeamName string
	Count    int
}

// ClaudeWatcherErrorMsg signals a watcher error
type ClaudeWatcherErrorMsg struct {
	Err error
}

// ClaudeWatcherReadyMsg signals that the claude dir exists and watcher can start
type ClaudeWatcherReadyMsg struct {
	ClaudeDir string
}
