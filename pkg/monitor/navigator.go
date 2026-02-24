package monitor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/Anansitrading/TLDR/internal/models"
)

// rebuildNavFlatItems rebuilds the flattened list of navigator items from the team tree.
func (m *Model) rebuildNavFlatItems() {
	nav := &m.Nav
	var items []NavItem

	// Calculate total count across all teams
	total := 0
	for _, t := range nav.Teams {
		total += t.Count
	}

	// "All Teams" header
	items = append(items, NavItem{
		Type:  NavItemAllTeams,
		Label: "All Teams",
		Count: total,
	})

	for _, team := range nav.Teams {
		teamLabel := team.Key
		if teamLabel == "" {
			teamLabel = "(Local)"
		}

		items = append(items, NavItem{
			Type:    NavItemTeam,
			TeamKey: team.Key,
			Label:   teamLabel,
			Count:   team.Count,
			Indent:  1,
		})

		// If expanded, show projects
		if nav.Expanded[team.Key] {
			// "All Projects" within this team (only if multiple projects)
			if len(team.Projects) > 1 {
				items = append(items, NavItem{
					Type:    NavItemAllProjects,
					TeamKey: team.Key,
					Label:   "All Projects",
					Count:   team.Count,
					Indent:  2,
				})
			}

			for _, proj := range team.Projects {
				projLabel := proj.Tag
				if projLabel == "" {
					projLabel = "(No Project)"
				}
				items = append(items, NavItem{
					Type:    NavItemProject,
					TeamKey: team.Key,
					Project: proj.Tag,
					Label:   projLabel,
					Count:   proj.Count,
					Indent:  2,
				})
			}
		}
	}

	nav.FlatItems = items

	// Clamp cursor
	if nav.Cursor >= len(items) {
		nav.Cursor = max(0, len(items)-1)
	}
}

// navSelect handles Enter on the current navigator item, applying the filter.
func (m *Model) navSelect() {
	nav := &m.Nav
	if nav.Cursor >= len(nav.FlatItems) {
		return
	}
	item := nav.FlatItems[nav.Cursor]

	switch item.Type {
	case NavItemAllTeams:
		nav.SelectedTeam = ""
		nav.SelectedProject = ""
	case NavItemTeam:
		// Toggle expand and select team
		nav.Expanded[item.TeamKey] = !nav.Expanded[item.TeamKey]
		nav.SelectedTeam = item.TeamKey
		nav.SelectedProject = ""
		m.rebuildNavFlatItems()
	case NavItemAllProjects:
		nav.SelectedTeam = item.TeamKey
		nav.SelectedProject = ""
	case NavItemProject:
		nav.SelectedTeam = item.TeamKey
		nav.SelectedProject = item.Project
	}
}

// navToggleExpand expands/collapses a team without changing selection
func (m *Model) navToggleExpand() {
	nav := &m.Nav
	if nav.Cursor >= len(nav.FlatItems) {
		return
	}
	item := nav.FlatItems[nav.Cursor]
	if item.Type == NavItemTeam {
		nav.Expanded[item.TeamKey] = !nav.Expanded[item.TeamKey]
		m.rebuildNavFlatItems()
	}
}

// filterActivityByTeamProject filters activity items based on the current team/project selection.
func (m *Model) filterActivityByTeamProject(items []ActivityItem) []ActivityItem {
	if m.Nav.SelectedTeam == "" && m.Nav.SelectedProject == "" {
		return items // No filter
	}

	var filtered []ActivityItem
	for _, item := range items {
		if item.IssueID == "" {
			filtered = append(filtered, item)
			continue
		}
		if m.matchesTeamProjectFilter(item.IssueID) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// filterBoardIssuesByTeamProject filters board issues based on the current team/project selection.
func (m *Model) filterBoardIssuesByTeamProject(issues []models.BoardIssueView) []models.BoardIssueView {
	if m.Nav.SelectedTeam == "" && m.Nav.SelectedProject == "" {
		return issues // No filter
	}

	var filtered []models.BoardIssueView
	for _, biv := range issues {
		teamKey := TeamKeyFromLinearID(biv.Issue.LinearIdentifier)
		if m.Nav.SelectedTeam != "" && teamKey != m.Nav.SelectedTeam {
			continue
		}
		if m.Nav.SelectedProject != "" && biv.Issue.ProjectTag != m.Nav.SelectedProject {
			continue
		}
		filtered = append(filtered, biv)
	}
	return filtered
}

// matchesTeamProjectFilter checks if an issue (by ID) matches the current team/project filter.
func (m *Model) matchesTeamProjectFilter(issueID string) bool {
	if m.Nav.SelectedTeam != "" {
		teamKey := m.IssueTeamMap[issueID]
		if teamKey != m.Nav.SelectedTeam {
			return false
		}
	}
	if m.Nav.SelectedProject != "" {
		projectTag := m.IssueProjectMap[issueID]
		if projectTag != m.Nav.SelectedProject {
			return false
		}
	}
	return true
}

// filterTaskListByTeamProject filters all categories in a TaskListData by the current
// team/project selection. Returns unmodified data if no filter is active.
func (m *Model) filterTaskListByTeamProject(data TaskListData) TaskListData {
	if m.Nav.SelectedTeam == "" && m.Nav.SelectedProject == "" {
		return data
	}
	filter := func(issues []models.Issue) []models.Issue {
		var filtered []models.Issue
		for _, issue := range issues {
			if m.matchesTeamProjectFilter(issue.ID) {
				filtered = append(filtered, issue)
			}
		}
		return filtered
	}
	return TaskListData{
		Ready:       filter(data.Ready),
		Reviewable:  filter(data.Reviewable),
		NeedsRework: filter(data.NeedsRework),
		Blocked:     filter(data.Blocked),
		Closed:      filter(data.Closed),
	}
}

// renderTeamProjectPanel renders the team/project navigator panel (Panel 1)
func (m Model) renderTeamProjectPanel(height int) string {
	nav := &m.Nav
	isActive := m.ActivePanel == PanelCurrentWork

	// Build title with filter indicator
	panelTitle := "TEAMS & PROJECTS"
	if nav.SelectedTeam != "" {
		teamLabel := nav.SelectedTeam
		if teamLabel == "" {
			teamLabel = "(Local)"
		}
		if nav.SelectedProject != "" {
			projLabel := nav.SelectedProject
			if projLabel == "" {
				projLabel = "(No Project)"
			}
			// Truncate long project names
			if len(projLabel) > 25 {
				projLabel = projLabel[:22] + "..."
			}
			panelTitle = fmt.Sprintf("FILTER: %s > %s", teamLabel, projLabel)
		} else {
			panelTitle = fmt.Sprintf("FILTER: %s", teamLabel)
		}
	}

	if len(nav.FlatItems) == 0 {
		content := subtleStyle.Render("No teams found")
		return m.wrapPanel(panelTitle, content+"\n", height, PanelCurrentWork)
	}

	var content strings.Builder
	maxLines := height - 3 // Panel border + title
	if maxLines < 1 {
		maxLines = 1
	}

	cursor := nav.Cursor
	offset := nav.ScrollOffset

	// Adjust scroll to keep cursor visible
	if cursor >= offset+maxLines {
		offset = cursor - maxLines + 1
	}
	if cursor < offset {
		offset = cursor
	}
	if offset < 0 {
		offset = 0
	}
	nav.ScrollOffset = offset

	contentWidth := m.Width - 4 // Panel border + padding

	// Scroll-up indicator
	if offset > 0 {
		content.WriteString(subtleStyle.Render("  ↑ more"))
		content.WriteString("\n")
		maxLines--
	}

	linesWritten := 0
	for i := offset; i < len(nav.FlatItems) && linesWritten < maxLines; i++ {
		item := nav.FlatItems[i]
		isSelected := isActive && i == cursor
		isActiveFilter := m.isActiveNavItem(item)

		// Build line with indentation
		indent := strings.Repeat("  ", item.Indent)
		var line string

		// Expand/collapse indicator for teams
		switch item.Type {
		case NavItemAllTeams:
			line = fmt.Sprintf("%s● All Teams (%d)", indent, item.Count)
		case NavItemTeam:
			arrow := "▸"
			if nav.Expanded[item.TeamKey] {
				arrow = "▾"
			}
			line = fmt.Sprintf("%s%s %s (%d)", indent, arrow, item.Label, item.Count)
		case NavItemAllProjects:
			line = fmt.Sprintf("%s  All Projects (%d)", indent, item.Count)
		case NavItemProject:
			line = fmt.Sprintf("%s  %s (%d)", indent, item.Label, item.Count)
		}

		// Truncate to panel width
		if lipgloss.Width(line) > contentWidth {
			line = ansi.Truncate(line, contentWidth-1, "…")
		}

		// Apply styling
		if isSelected {
			line = highlightRow(line, contentWidth)
		} else if isActiveFilter {
			// Highlight the currently active filter item
			activeStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214")) // Orange for active filter
			padLen := contentWidth - lipgloss.Width(line)
			if padLen > 0 {
				line += strings.Repeat(" ", padLen)
			}
			line = activeStyle.Render(line)
		}

		content.WriteString(line)
		content.WriteString("\n")
		linesWritten++
	}

	// Scroll-down indicator
	remaining := len(nav.FlatItems) - offset - maxLines
	if remaining > 0 {
		content.WriteString(subtleStyle.Render(fmt.Sprintf("  ↓ %d more", remaining)))
		content.WriteString("\n")
	}

	return m.wrapPanel(panelTitle, content.String(), height, PanelCurrentWork)
}

// isActiveNavItem checks if a nav item represents the currently active filter.
func (m Model) isActiveNavItem(item NavItem) bool {
	switch item.Type {
	case NavItemAllTeams:
		return m.Nav.SelectedTeam == "" && m.Nav.SelectedProject == ""
	case NavItemTeam:
		return m.Nav.SelectedTeam == item.TeamKey && m.Nav.SelectedProject == ""
	case NavItemAllProjects:
		return m.Nav.SelectedTeam == item.TeamKey && m.Nav.SelectedProject == ""
	case NavItemProject:
		return m.Nav.SelectedTeam == item.TeamKey && m.Nav.SelectedProject == item.Project
	}
	return false
}
