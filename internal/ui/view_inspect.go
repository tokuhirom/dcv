package ui

import (
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InspectViewModel struct {
	SearchViewModel

	// Inspect view state
	inspectContent string
	inspectScrollY int

	inspectTargetName string
}

// render renders the container inspect view
func (m *InspectViewModel) render(availableHeight int) string {
	var content strings.Builder

	// Split content into lines for scrolling
	lines := strings.Split(m.inspectContent, "\n")
	viewHeight := availableHeight - 1
	startIdx := m.inspectScrollY
	endIdx := startIdx + viewHeight

	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	// Line number style
	lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("76"))

	// Define highlight style for search matches
	highlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("226")).
		Foreground(lipgloss.Color("235"))

	if len(lines) == 0 {
		content.WriteString("No inspect data available\n")
	} else if startIdx < len(lines) {
		for i := startIdx; i < endIdx; i++ {
			line := lines[i]
			lineNum := lineNumStyle.Render(fmt.Sprintf("%4d ", i+1))

			// Mark current search result line
			if len(m.searchResults) > 0 && m.currentSearchIdx < len(m.searchResults) &&
				i == m.searchResults[m.currentSearchIdx] {
				// Replace the first character with a styled marker, avoiding string slicing of styled text
				marker := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("▶")
				lineNumPlain := fmt.Sprintf("%4d ", i+1)
				lineNum = marker + lineNumStyle.Render(lineNumPlain[1:])
			}

			// Apply both YAML syntax highlighting and search highlighting
			highlightedLine := m.renderLineWithHighlighting(line, keyStyle, valueStyle, highlightStyle)

			content.WriteString(lineNum + highlightedLine + "\n")
		}
	}

	// Fill remaining space
	linesShown := endIdx - startIdx
	for i := linesShown; i < viewHeight; i++ {
		content.WriteString("\n")
	}

	// Show position indicator
	if len(lines) > viewHeight {
		posStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		position := fmt.Sprintf("Lines %d-%d of %d", startIdx+1, endIdx, len(lines))
		if len(m.searchResults) > 0 {
			position += fmt.Sprintf(" | Match %d/%d", m.currentSearchIdx+1, len(m.searchResults))
		}
		content.WriteString(posStyle.Render(position))
	}

	return content.String()
}

// renderLineWithHighlighting applies both YAML syntax highlighting and search highlighting
func (m *InspectViewModel) renderLineWithHighlighting(line string, keyStyle, valueStyle, highlightStyle lipgloss.Style) string {
	// If no search is active, just apply YAML highlighting
	if m.searchText == "" || m.searchMode {
		return m.applyYAMLHighlighting(line, keyStyle, valueStyle)
	}

	// Apply search highlighting with preserved YAML styling
	return m.applySearchWithYAMLHighlighting(line, keyStyle, valueStyle, highlightStyle)
}

// applyYAMLHighlighting applies YAML syntax highlighting to a line
func (m *InspectViewModel) applyYAMLHighlighting(line string, keyStyle, valueStyle lipgloss.Style) string {
	trimmed := strings.TrimSpace(line)

	// Check if this line is a YAML key-value pair
	if idx := strings.Index(line, ": "); idx != -1 && !strings.HasPrefix(trimmed, "-") {
		// Split into key and value parts
		keyPart := line[:idx]
		valuePart := line[idx+2:]
		return keyStyle.Render(keyPart) + ": " + valueStyle.Render(valuePart)
	} else if strings.HasPrefix(trimmed, "- ") {
		// YAML list item
		indent := line[:len(line)-len(trimmed)]
		content := trimmed[2:] // Remove "- "
		return indent + "- " + valueStyle.Render(content)
	} else if trimmed != "" {
		// Other content
		return valueStyle.Render(line)
	}

	return line
}

// applySearchWithYAMLHighlighting applies search highlighting while preserving YAML syntax colors
func (m *InspectViewModel) applySearchWithYAMLHighlighting(line string, keyStyle, valueStyle, highlightStyle lipgloss.Style) string {
	trimmed := strings.TrimSpace(line)

	// Find search matches first
	searchMatches := m.findSearchMatches(line)
	if len(searchMatches) == 0 {
		return m.applyYAMLHighlighting(line, keyStyle, valueStyle)
	}

	// Build the result by applying both highlighting types
	var result strings.Builder

	// Check if this line is a YAML key-value pair
	if idx := strings.Index(line, ": "); idx != -1 && !strings.HasPrefix(trimmed, "-") {
		// Handle key-value pair with search highlighting
		keyPart := line[:idx]
		separatorPart := line[idx : idx+2] // ": "
		valuePart := line[idx+2:]

		// Apply highlighting to each part
		result.WriteString(m.applyHighlightingToPart(keyPart, keyStyle, highlightStyle, searchMatches, 0))
		result.WriteString(separatorPart)
		result.WriteString(m.applyHighlightingToPart(valuePart, valueStyle, highlightStyle, searchMatches, idx+2))

	} else if strings.HasPrefix(trimmed, "- ") {
		// YAML list item
		indent := line[:len(line)-len(trimmed)]
		prefix := trimmed[:2] // "- "
		content := trimmed[2:]
		prefixStart := len(indent)
		contentStart := prefixStart + 2

		result.WriteString(indent)
		result.WriteString(prefix)
		result.WriteString(m.applyHighlightingToPart(content, valueStyle, highlightStyle, searchMatches, contentStart))

	} else {
		// Other content
		result.WriteString(m.applyHighlightingToPart(line, valueStyle, highlightStyle, searchMatches, 0))
	}

	return result.String()
}

// findSearchMatches finds all search match positions in a line
func (m *InspectViewModel) findSearchMatches(line string) [][]int {
	var matches [][]int

	if m.searchRegex {
		pattern := m.searchText
		if m.searchIgnoreCase {
			pattern = "(?i)" + pattern
		}
		if re, err := regexp.Compile(pattern); err == nil {
			matches = re.FindAllStringIndex(line, -1)
		}
	} else {
		searchStr := m.searchText
		lineToSearch := line

		if m.searchIgnoreCase {
			searchStr = strings.ToLower(searchStr)
			lineToSearch = strings.ToLower(lineToSearch)
		}

		start := 0
		for {
			idx := strings.Index(lineToSearch[start:], searchStr)
			if idx == -1 {
				break
			}
			realIdx := start + idx
			matches = append(matches, []int{realIdx, realIdx + len(searchStr)})
			start = realIdx + 1
		}
	}

	return matches
}

// applyHighlightingToPart applies highlighting to a part of the line
func (m *InspectViewModel) applyHighlightingToPart(part string, baseStyle, highlightStyle lipgloss.Style, allMatches [][]int, partOffset int) string {
	if len(allMatches) == 0 {
		return baseStyle.Render(part)
	}

	// Find matches that overlap with this part
	var partMatches [][]int
	for _, match := range allMatches {
		start, end := match[0], match[1]

		// Adjust match positions relative to the part
		relStart := start - partOffset
		relEnd := end - partOffset

		// Check if the match overlaps with this part
		if relEnd > 0 && relStart < len(part) {
			// Clamp to part boundaries
			if relStart < 0 {
				relStart = 0
			}
			if relEnd > len(part) {
				relEnd = len(part)
			}
			partMatches = append(partMatches, []int{relStart, relEnd})
		}
	}

	if len(partMatches) == 0 {
		return baseStyle.Render(part)
	}

	// Apply highlighting to the part
	var result strings.Builder
	lastEnd := 0

	for _, match := range partMatches {
		start, end := match[0], match[1]

		// Add non-highlighted text before the match
		if start > lastEnd {
			result.WriteString(baseStyle.Render(part[lastEnd:start]))
		}

		// Add highlighted match
		result.WriteString(highlightStyle.Render(part[start:end]))
		lastEnd = end
	}

	// Add remaining non-highlighted text
	if lastEnd < len(part) {
		result.WriteString(baseStyle.Render(part[lastEnd:]))
	}

	return result.String()
}

func (m *InspectViewModel) HandleBack(model *Model) tea.Cmd {
	// ClearSearch search state when leaving inspect view
	m.searchMode = false
	m.searchText = ""
	m.searchResults = nil
	m.currentSearchIdx = 0

	model.SwitchToPreviousView()

	return nil
}

func (m *InspectViewModel) HandleUp() tea.Cmd {
	if m.inspectScrollY > 0 {
		m.inspectScrollY--
	}
	return nil
}

func (m *InspectViewModel) HandleDown(model *Model) tea.Cmd {
	lines := strings.Split(m.inspectContent, "\n")
	maxScroll := len(lines) - (model.Height - 5)
	if m.inspectScrollY < maxScroll && maxScroll > 0 {
		m.inspectScrollY++
	}
	return nil
}

func (m *InspectViewModel) HandleGoToEnd(model *Model) tea.Cmd {
	lines := strings.Split(m.inspectContent, "\n")
	maxScroll := len(lines) - (model.Height - 5)
	if maxScroll > 0 {
		m.inspectScrollY = maxScroll
	}
	return nil
}

func (m *InspectViewModel) HandleGoToStart() tea.Cmd {
	m.inspectScrollY = 0
	return nil
}

func (m *InspectViewModel) HandlePageUp(model *Model) tea.Cmd {
	pageSize := model.Height - 5
	m.inspectScrollY -= pageSize
	if m.inspectScrollY < 0 {
		m.inspectScrollY = 0
	}
	return nil
}

func (m *InspectViewModel) HandlePageDown(model *Model) tea.Cmd {
	pageSize := model.Height - 5
	lines := strings.Split(m.inspectContent, "\n")
	maxScroll := len(lines) - pageSize

	m.inspectScrollY += pageSize
	if m.inspectScrollY > maxScroll && maxScroll > 0 {
		m.inspectScrollY = maxScroll
	} else if maxScroll <= 0 {
		m.inspectScrollY = 0
	}
	return nil
}

func (m *InspectViewModel) HandleSearch() tea.Cmd {
	m.ClearSearch()
	return nil
}

func (m *InspectViewModel) HandleNextSearchResult(model *Model) tea.Cmd {
	if len(m.searchResults) > 0 {
		m.currentSearchIdx = (m.currentSearchIdx + 1) % len(m.searchResults)
		// Jump to the line
		if m.currentSearchIdx < len(m.searchResults) {
			targetLine := m.searchResults[m.currentSearchIdx]
			m.inspectScrollY = targetLine - model.Height/2 + 3 // Center the result
			if m.inspectScrollY < 0 {
				m.inspectScrollY = 0
			}
		}
	}
	return nil
}

func (m *InspectViewModel) HandlePrevSearchResult(model *Model) tea.Cmd {
	if len(m.searchResults) > 0 {
		m.currentSearchIdx--
		if m.currentSearchIdx < 0 {
			m.currentSearchIdx = len(m.searchResults) - 1
		}
		// Jump to the line
		if m.currentSearchIdx < len(m.searchResults) {
			targetLine := m.searchResults[m.currentSearchIdx]
			m.inspectScrollY = targetLine - model.Height/2 + 3 // Center the result
			if m.inspectScrollY < 0 {
				m.inspectScrollY = 0
			}
		}
	}
	return nil
}

func (m *InspectViewModel) Set(content string, targetName string) {
	m.inspectContent = content
	m.inspectTargetName = targetName
	m.inspectScrollY = 0
}

func (m *InspectViewModel) Title() string {
	base := "Inspect " + m.inspectTargetName + " "

	// Add search status if applicable
	if m.searchText != "" && !m.searchMode {
		searchInfo := fmt.Sprintf(" | Search: %s", m.searchText)
		if len(m.searchResults) > 0 {
			searchInfo += fmt.Sprintf(" (%d/%d)", m.currentSearchIdx+1, len(m.searchResults))
		} else {
			searchInfo += " (no matches)"
		}
		if m.searchIgnoreCase {
			searchInfo += " [i]"
		}
		if m.searchRegex {
			searchInfo += " [re]"
		}
		base += searchInfo
	}
	return base
}

type InspectProvider func() ([]byte, error)

func (m *InspectViewModel) Inspect(model *Model, targetName string, inspectProvider InspectProvider) tea.Cmd {
	model.loading = true
	return func() tea.Msg {
		content, err := inspectProvider()
		return inspectLoadedMsg{
			content:    string(content),
			err:        err,
			targetName: targetName,
		}
	}
}
