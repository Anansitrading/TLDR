package monitor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/marcus/td/internal/models"
)

// kanbanStatusOrder defines the display order for kanban columns
var kanbanStatusOrder = []models.Status{
	models.StatusOpen,
	models.StatusInProgress,
	models.StatusInReview,
	models.StatusBlocked,
	models.StatusClosed,
}

// rebuildKanbanState rebuilds the kanban column data from current board issues.
// Called when switching to kanban view or when board data refreshes.
func (m *Model) rebuildKanbanState() {
	ks := &m.BoardMode.Kanban

	// Initialize maps if nil
	if ks.ColumnCursors == nil {
		ks.ColumnCursors = make(map[models.Status]int)
	}
	if ks.ColumnScrolls == nil {
		ks.ColumnScrolls = make(map[models.Status]int)
	}
	if ks.ColumnIssues == nil {
		ks.ColumnIssues = make(map[models.Status][]models.Issue)
	}

	// Clear column issues
	for k := range ks.ColumnIssues {
		delete(ks.ColumnIssues, k)
	}

	// Group issues by status
	for _, biv := range m.BoardMode.Issues {
		status := biv.Issue.Status
		// Check status filter
		if visible, ok := m.BoardMode.StatusFilter[status]; ok && !visible {
			continue
		}
		ks.ColumnIssues[status] = append(ks.ColumnIssues[status], biv.Issue)
	}

	// Build visible statuses in canonical order (skip empty ones)
	ks.VisibleStatuses = nil
	for _, status := range kanbanStatusOrder {
		if issues, ok := ks.ColumnIssues[status]; ok && len(issues) > 0 {
			ks.VisibleStatuses = append(ks.VisibleStatuses, status)
		}
	}

	// Clamp active column
	if ks.ActiveColumn >= len(ks.VisibleStatuses) {
		ks.ActiveColumn = max(0, len(ks.VisibleStatuses)-1)
	}

	// Clamp cursors for each column
	for status, issues := range ks.ColumnIssues {
		if cursor, ok := ks.ColumnCursors[status]; ok {
			if cursor >= len(issues) {
				ks.ColumnCursors[status] = max(0, len(issues)-1)
			}
		}
		if scroll, ok := ks.ColumnScrolls[status]; ok {
			if scroll >= len(issues) {
				ks.ColumnScrolls[status] = max(0, len(issues)-1)
			}
		}
	}
}

// kanbanCursorDown moves the cursor down in the active kanban column
func (m *Model) kanbanCursorDown() {
	ks := &m.BoardMode.Kanban
	if len(ks.VisibleStatuses) == 0 {
		return
	}
	status := ks.VisibleStatuses[ks.ActiveColumn]
	issues := ks.ColumnIssues[status]
	cursor := ks.ColumnCursors[status]
	if cursor < len(issues)-1 {
		ks.ColumnCursors[status] = cursor + 1
	}
}

// kanbanCursorUp moves the cursor up in the active kanban column
func (m *Model) kanbanCursorUp() {
	ks := &m.BoardMode.Kanban
	if len(ks.VisibleStatuses) == 0 {
		return
	}
	status := ks.VisibleStatuses[ks.ActiveColumn]
	cursor := ks.ColumnCursors[status]
	if cursor > 0 {
		ks.ColumnCursors[status] = cursor - 1
	}
}

// kanbanCursorTop jumps to the top of the active kanban column
func (m *Model) kanbanCursorTop() {
	ks := &m.BoardMode.Kanban
	if len(ks.VisibleStatuses) == 0 {
		return
	}
	status := ks.VisibleStatuses[ks.ActiveColumn]
	ks.ColumnCursors[status] = 0
	ks.ColumnScrolls[status] = 0
}

// kanbanCursorBottom jumps to the bottom of the active kanban column
func (m *Model) kanbanCursorBottom() {
	ks := &m.BoardMode.Kanban
	if len(ks.VisibleStatuses) == 0 {
		return
	}
	status := ks.VisibleStatuses[ks.ActiveColumn]
	issues := ks.ColumnIssues[status]
	if len(issues) > 0 {
		ks.ColumnCursors[status] = len(issues) - 1
	}
}

// kanbanPrevColumn moves to the previous kanban column
func (m *Model) kanbanPrevColumn() {
	ks := &m.BoardMode.Kanban
	if ks.ActiveColumn > 0 {
		ks.ActiveColumn--
	}
}

// kanbanNextColumn moves to the next kanban column
func (m *Model) kanbanNextColumn() {
	ks := &m.BoardMode.Kanban
	if ks.ActiveColumn < len(ks.VisibleStatuses)-1 {
		ks.ActiveColumn++
	}
}

// kanbanPageMove moves the cursor in the active column by delta (half/full page)
func (m *Model) kanbanPageMove(delta int) {
	ks := &m.BoardMode.Kanban
	if len(ks.VisibleStatuses) == 0 {
		return
	}
	status := ks.VisibleStatuses[ks.ActiveColumn]
	issues := ks.ColumnIssues[status]
	if len(issues) == 0 {
		return
	}
	cursor := ks.ColumnCursors[status] + delta
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= len(issues) {
		cursor = len(issues) - 1
	}
	ks.ColumnCursors[status] = cursor
}

// indexOf returns the index of a status in a slice, or 0 if not found
func indexOf(status models.Status, statuses []models.Status) int {
	for i, s := range statuses {
		if s == status {
			return i
		}
	}
	return 0
}

// renderBoardKanbanView renders the kanban board view with side-by-side status columns
func (m Model) renderBoardKanbanView(height int) string {
	ks := &m.BoardMode.Kanban
	boardName := "Board"
	if m.BoardMode.Board != nil {
		boardName = m.BoardMode.Board.Name
	}

	// Build sort indicator
	sortIndicator := ""
	switch m.SortMode {
	case SortByCreatedDesc:
		sortIndicator = " [by:created]"
	case SortByUpdatedDesc:
		sortIndicator = " [by:updated]"
	}

	// Count total issues across all columns
	totalIssues := 0
	for _, issues := range ks.ColumnIssues {
		totalIssues += len(issues)
	}

	panelTitle := fmt.Sprintf("BOARD: %s [kanban]%s (%d)", boardName, sortIndicator, totalIssues)

	if len(ks.VisibleStatuses) == 0 {
		content := subtleStyle.Render("No issues match the board query")
		content += "\n\n"
		content += subtleStyle.Render("Try adjusting the status filter with 'c' or 'F'")
		return m.wrapPanel(panelTitle, content, height, PanelTaskList)
	}

	isActive := m.ActivePanel == PanelTaskList
	numCols := len(ks.VisibleStatuses)

	// Calculate available width for columns (panel content width minus borders)
	panelContentWidth := m.Width - 4 // Panel border + padding
	if panelContentWidth < numCols*8 {
		panelContentWidth = numCols * 8
	}

	// Column width: divide equally, give remainder to leftmost columns
	colWidth := panelContentWidth / numCols
	extraPixels := panelContentWidth - (colWidth * numCols)

	// Available height for card content (subtract panel border + header)
	cardAreaHeight := height - 5 // Panel border (2) + title (1) + column header (1) + bottom padding (1)
	if cardAreaHeight < 3 {
		cardAreaHeight = 3
	}

	// Render each column
	columns := make([]string, numCols)
	for i, status := range ks.VisibleStatuses {
		issues := ks.ColumnIssues[status]
		cursor := ks.ColumnCursors[status]
		isActiveCol := isActive && i == ks.ActiveColumn

		// Determine this column's width (add 1 to first `extraPixels` columns)
		thisColWidth := colWidth
		if i < extraPixels {
			thisColWidth++
		}

		columns[i] = m.renderKanbanColumn(status, issues, cursor, thisColWidth, cardAreaHeight, isActiveCol)
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	return m.wrapPanel(panelTitle, content, height, PanelTaskList)
}

// renderKanbanColumn renders a single kanban column with header and cards
func (m Model) renderKanbanColumn(status models.Status, issues []models.Issue, cursor int, width int, maxCardHeight int, isActive bool) string {
	var sb strings.Builder

	// Column header: status name + count with colored background
	statusColor, ok := kanbanStatusColors[status]
	if !ok {
		statusColor = lipgloss.Color("241")
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("0")). // Black text
		Background(statusColor).
		Width(width - 2). // Subtract border width
		Align(lipgloss.Center)

	statusName := strings.ToUpper(string(status))
	headerText := fmt.Sprintf(" %s (%d) ", statusName, len(issues))
	sb.WriteString(headerStyle.Render(headerText))
	sb.WriteString("\n")

	if len(issues) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(width - 2).
			Align(lipgloss.Center)
		sb.WriteString(emptyStyle.Render("empty"))
		// Pad to fill height
		for i := 1; i < maxCardHeight; i++ {
			sb.WriteString("\n")
		}
	} else {
		// Calculate how many cards fit (each card is 2 lines + 1 separator = 3 lines, except last)
		linesPerCard := 3
		visibleCards := maxCardHeight / linesPerCard
		if visibleCards < 1 {
			visibleCards = 1
		}

		// Handle scroll offset
		scroll := m.BoardMode.Kanban.ColumnScrolls[status]
		if cursor >= scroll+visibleCards {
			scroll = cursor - visibleCards + 1
		}
		if cursor < scroll {
			scroll = cursor
		}
		if scroll < 0 {
			scroll = 0
		}
		m.BoardMode.Kanban.ColumnScrolls[status] = scroll

		// Render visible cards
		linesUsed := 0
		contentWidth := width - 2 // Subtract column border

		// Show scroll-up indicator
		if scroll > 0 {
			upText := fmt.Sprintf("  ↑ %d more", scroll)
			sb.WriteString(subtleStyle.Render(upText))
			sb.WriteString("\n")
			linesUsed++
		}

		for idx := scroll; idx < len(issues) && linesUsed < maxCardHeight; idx++ {
			issue := issues[idx]
			isSelected := isActive && idx == cursor

			// Card line 1: TypeIcon ID Priority
			typeIcon := formatTypeIcon(issue.Type)
			priority := formatPriority(issue.Priority)
			idStr := issue.ID
			if len(idStr) > 10 {
				idStr = idStr[:10]
			}
			line1 := fmt.Sprintf("%s %s %s", typeIcon, idStr, priority)

			// Card line 2: truncated title
			title := issue.Title
			titleWidth := contentWidth - 1
			if titleWidth < 4 {
				titleWidth = 4
			}
			if lipgloss.Width(title) > titleWidth {
				title = ansi.Truncate(title, titleWidth-1, "…")
			}
			line2 := " " + title

			if isSelected {
				line1 = highlightRow(line1, contentWidth)
				line2 = highlightRow(line2, contentWidth)
			} else {
				// Pad lines to column width
				line1Pad := contentWidth - lipgloss.Width(line1)
				if line1Pad > 0 {
					line1 += strings.Repeat(" ", line1Pad)
				}
				line2Pad := contentWidth - lipgloss.Width(line2)
				if line2Pad > 0 {
					line2 += strings.Repeat(" ", line2Pad)
				}
			}

			sb.WriteString(line1)
			sb.WriteString("\n")
			linesUsed++

			if linesUsed < maxCardHeight {
				sb.WriteString(line2)
				sb.WriteString("\n")
				linesUsed++
			}

			// Add separator between cards (not after last card)
			if linesUsed < maxCardHeight && idx < len(issues)-1 {
				sep := subtleStyle.Render(strings.Repeat(kanbanCardSeparator, contentWidth/len(kanbanCardSeparator)))
				sb.WriteString(sep)
				sb.WriteString("\n")
				linesUsed++
			}
		}

		// Show scroll-down indicator
		remaining := len(issues) - scroll - visibleCards
		if remaining > 0 {
			if linesUsed < maxCardHeight {
				downIndicator := subtleStyle.Render(fmt.Sprintf("  ↓ %d more", remaining))
				sb.WriteString(downIndicator)
				sb.WriteString("\n")
				linesUsed++
			}
		}

		// Pad remaining space
		for linesUsed < maxCardHeight {
			sb.WriteString("\n")
			linesUsed++
		}
	}

	// Wrap in column border
	content := sb.String()
	var colStyle lipgloss.Style
	if isActive {
		colStyle = kanbanActiveColumnStyle.
			Width(width).
			Height(maxCardHeight + 1). // +1 for header
			BorderForeground(statusColor)
	} else {
		colStyle = kanbanColumnStyle.
			Width(width).
			Height(maxCardHeight + 1)
	}

	return colStyle.Render(content)
}
