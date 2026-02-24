package claudewatch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/Anansitrading/TLDR/internal/db"
)

// Watcher monitors Claude Code task directories and imports tasks into td.
type Watcher struct {
	claudeDir string
	database  *db.DB
	notify    func(teamName string, count int)
	mu        sync.Mutex
}

// NewWatcher creates a new Claude Code task watcher.
func NewWatcher(claudeDir string, database *db.DB) *Watcher {
	return &Watcher{
		claudeDir: claudeDir,
		database:  database,
	}
}

// SetNotify sets the callback invoked when tasks are imported or updated.
func (w *Watcher) SetNotify(fn func(teamName string, count int)) {
	w.notify = fn
}

// ScanAndImport scans all Claude Code task directories and imports/updates tasks.
// Returns the number of tasks successfully imported or updated.
func (w *Watcher) ScanAndImport() (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	tasksDir := filepath.Join(w.claudeDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read tasks dir: %w", err)
	}

	total := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		teamName := entry.Name()
		count, err := w.scanTeam(teamName)
		if err != nil {
			// Skip teams that fail, continue with others
			continue
		}
		total += count
	}
	return total, nil
}

// scanTeam reads all task JSON files for a single team and upserts them.
func (w *Watcher) scanTeam(teamName string) (int, error) {
	teamTaskDir := filepath.Join(w.claudeDir, "tasks", teamName)
	entries, err := os.ReadDir(teamTaskDir)
	if err != nil {
		return 0, err
	}

	projectTag := ""
	proj, err := w.database.GetProjectByClaudeTeam(teamName)
	if err == nil && proj != nil {
		projectTag = proj.Name
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(teamTaskDir, entry.Name()))
		if err != nil {
			continue
		}
		var task ClaudeTask
		if err := json.Unmarshal(data, &task); err != nil {
			continue
		}
		if task.Subject == "" {
			continue
		}
		if err := w.upsertTask(&task, teamName, projectTag); err != nil {
			continue
		}
		count++
	}
	return count, nil
}

// upsertTask creates or updates a td issue from a Claude Code task.
func (w *Watcher) upsertTask(task *ClaudeTask, teamName, projectTag string) error {
	linearID := MakeLinearID(teamName, task.ID)

	existing, err := w.database.GetIssueByLinearID(linearID)
	if err == nil && existing != nil {
		// Update existing issue
		existing.Title = task.Subject
		existing.Description = task.Description
		existing.Status = MapStatus(task.Status)
		existing.Assignee = task.Owner
		if projectTag != "" {
			existing.ProjectTag = projectTag
		}
		return w.database.UpdateIssue(existing)
	}

	// Create new issue, then update to set watcher-specific fields
	// (CreateIssue doesn't include assignee/linear_id/project_tag in INSERT)
	issue := MapToIssue(task, teamName, projectTag)
	if err := w.database.CreateIssue(issue); err != nil {
		return fmt.Errorf("create issue from claude task: %w", err)
	}
	return w.database.UpdateIssue(issue)
}

// StartWatch begins watching Claude Code directories for changes using fsnotify + polling.
func (w *Watcher) StartWatch(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create fsnotify watcher: %w", err)
	}

	tasksDir := filepath.Join(w.claudeDir, "tasks")
	os.MkdirAll(tasksDir, 0755)
	if err := watcher.Add(tasksDir); err != nil {
		watcher.Close()
		return fmt.Errorf("watch tasks dir: %w", err)
	}

	// Watch existing team subdirectories
	entries, _ := os.ReadDir(tasksDir)
	for _, entry := range entries {
		if entry.IsDir() {
			watcher.Add(filepath.Join(tasksDir, entry.Name()))
		}
	}

	// Initial scan
	w.ScanAndImport()

	// Event loop goroutine
	go func() {
		defer watcher.Close()
		pollTicker := time.NewTicker(10 * time.Second)
		defer pollTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) || event.Has(fsnotify.Remove) {
					// Auto-watch new subdirectories
					if event.Has(fsnotify.Create) {
						if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
							watcher.Add(event.Name)
						}
					}
					time.Sleep(100 * time.Millisecond) // debounce
					count, _ := w.ScanAndImport()
					if count > 0 && w.notify != nil {
						teamName := w.teamNameFromPath(event.Name)
						w.notify(teamName, count)
					}
				}

			case <-pollTicker.C:
				count, _ := w.ScanAndImport()
				if count > 0 && w.notify != nil {
					w.notify("", count)
				}

			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	return nil
}

// teamNameFromPath extracts the team name from a file path under the tasks directory.
func (w *Watcher) teamNameFromPath(path string) string {
	tasksDir := filepath.Join(w.claudeDir, "tasks")
	rel, err := filepath.Rel(tasksDir, path)
	if err != nil {
		return ""
	}
	parts := strings.SplitN(rel, string(filepath.Separator), 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
