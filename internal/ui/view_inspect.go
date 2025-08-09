package ui

import (
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/models"
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

	// Render the JSON content with syntax highlighting
	lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	jsonKeyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	jsonValueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("76"))
	jsonBraceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

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
				// Add a marker in the margin
				lineNum = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("▶") + lineNum[1:]
			}

			// Apply JSON syntax highlighting first
			highlightedLine := line
			if strings.Contains(line, "\":") {
				// This line likely contains a key-value pair
				parts := strings.SplitN(line, "\":", 2)
				if len(parts) == 2 {
					// Extract key part
					keyStart := strings.LastIndex(parts[0], "\"")
					if keyStart >= 0 {
						indent := parts[0][:keyStart]
						key := parts[0][keyStart:]
						value := parts[1]

						// Apply styles
						highlightedLine = indent + jsonKeyStyle.Render(key+"\":") + jsonValueStyle.Render(value)
					}
				}
			} else if strings.TrimSpace(line) == "{" || strings.TrimSpace(line) == "}" ||
				strings.TrimSpace(line) == "[" || strings.TrimSpace(line) == "]" ||
				strings.TrimSpace(line) == "}," || strings.TrimSpace(line) == "]," {
				// Highlight braces and brackets
				highlightedLine = jsonBraceStyle.Render(line)
			}

			// Then apply search highlighting if in search mode and have search text
			if m.searchText != "" && !m.searchMode {
				highlightedLine = m.highlightInspectLine(line, highlightedLine, highlightStyle)
			}

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

// HighlightInspectLine highlights search matches in a line that may already have JSON syntax highlighting
func (m *InspectViewModel) highlightInspectLine(originalLine, styledLine string, highlightStyle lipgloss.Style) string {
	if m.searchText == "" {
		return styledLine
	}

	// For inspect view, we need to be careful not to break the existing JSON syntax highlighting
	// We'll use a simpler approach - just highlight in the original line for display
	if m.searchRegex {
		pattern := m.searchText
		if m.searchIgnoreCase {
			pattern = "(?i)" + pattern
		}
		if re, err := regexp.Compile(pattern); err == nil {
			return re.ReplaceAllStringFunc(originalLine, func(match string) string {
				return highlightStyle.Render(match)
			})
		}
	} else {
		// Simple string search
		searchStr := m.searchText
		lineToSearch := originalLine

		if m.searchIgnoreCase {
			searchStr = strings.ToLower(searchStr)
			lineToSearch = strings.ToLower(originalLine)
		}

		// Find all occurrences
		var result strings.Builder
		lastEnd := 0
		for {
			idx := strings.Index(lineToSearch[lastEnd:], searchStr)
			if idx == -1 {
				break
			}

			realIdx := lastEnd + idx
			result.WriteString(originalLine[lastEnd:realIdx])
			result.WriteString(highlightStyle.Render(originalLine[realIdx : realIdx+len(m.searchText)]))
			lastEnd = realIdx + len(m.searchText)
		}
		result.WriteString(originalLine[lastEnd:])
		return result.String()
	}

	return styledLine
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

func (m *InspectViewModel) InspectImage(model *Model, image models.DockerImage, targetName string) tea.Cmd {
	model.loading = true
	return func() tea.Msg {
		content, err := model.dockerClient.ExecuteCaptured("image", "inspect", image.ID)
		return inspectLoadedMsg{
			content:    string(content),
			targetName: targetName,
			err:        err,
		}
	}
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

func (m *InspectViewModel) InspectVolume(model *Model, volume models.DockerVolume) tea.Cmd {
	model.SwitchView(InspectView)
	return func() tea.Msg {
		content, err := model.dockerClient.ExecuteCaptured("volume", "inspect", volume.Name)
		return inspectLoadedMsg{
			content:    string(content),
			err:        err,
			targetName: "volume " + volume.Name,
		}
	}
}
