package views

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"

	"github.com/tokuhirom/dcv/internal/docker"
)

// InspectView displays detailed information about Docker resources
type InspectView struct {
	docker           *docker.Client
	textView         *tview.TextView
	content          string
	targetName       string
	targetType       string // "container", "image", "volume", "network"
	scrollY          int
	maxScrollY       int
	searchText       string
	searchRegex      bool
	searchIgnoreCase bool
	searchResults    []int // Line numbers containing matches
	currentSearchIdx int
	lines            []string
	isSearchMode     bool
	searchInput      *tview.InputField
}

// NewInspectView creates a new inspect view
func NewInspectView(dockerClient *docker.Client) *InspectView {
	v := &InspectView{
		docker:      dockerClient,
		textView:    tview.NewTextView(),
		searchInput: tview.NewInputField(),
	}

	v.setupTextView()
	v.setupSearchInput()

	return v
}

// setupTextView configures the text view widget
func (v *InspectView) setupTextView() {
	v.textView.
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(false).
		SetChangedFunc(func() {
			// Update scroll position when content changes
			row, _ := v.textView.GetScrollOffset()
			v.scrollY = row
		})

	// Set up key handlers
	v.textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if v.isSearchMode {
			// In search mode, pass events to search input
			return event
		}

		switch event.Rune() {
		case 'j':
			// Scroll down (vim style)
			row, col := v.textView.GetScrollOffset()
			v.textView.ScrollTo(row+1, col)
			return nil

		case 'k':
			// Scroll up (vim style)
			row, col := v.textView.GetScrollOffset()
			if row > 0 {
				v.textView.ScrollTo(row-1, col)
			}
			return nil

		case 'g':
			// Go to top (vim style)
			v.textView.ScrollTo(0, 0)
			return nil

		case 'G':
			// Go to bottom (vim style)
			v.textView.ScrollToEnd()
			return nil

		case 'h':
			// Scroll left
			row, col := v.textView.GetScrollOffset()
			if col > 0 {
				v.textView.ScrollTo(row, col-1)
			}
			return nil

		case 'l':
			// Scroll right
			row, col := v.textView.GetScrollOffset()
			v.textView.ScrollTo(row, col+1)
			return nil

		case '/':
			// Start search
			v.startSearch()
			return nil

		case 'n':
			// Next search result
			v.nextSearchResult()
			return nil

		case 'N':
			// Previous search result
			v.prevSearchResult()
			return nil

		case 'y':
			// Yank (copy) current line
			// TODO: Implement clipboard support
			slog.Info("Yank line (clipboard support not implemented)")
			return nil

		case 'Y':
			// Yank (copy) entire content
			// TODO: Implement clipboard support
			slog.Info("Yank all (clipboard support not implemented)")
			return nil
		}

		switch event.Key() {
		case tcell.KeyPgUp:
			// Page up
			row, col := v.textView.GetScrollOffset()
			_, _, _, height := v.textView.GetInnerRect()
			newRow := row - height
			if newRow < 0 {
				newRow = 0
			}
			v.textView.ScrollTo(newRow, col)
			return nil

		case tcell.KeyPgDn:
			// Page down
			row, col := v.textView.GetScrollOffset()
			_, _, _, height := v.textView.GetInnerRect()
			v.textView.ScrollTo(row+height, col)
			return nil

		case tcell.KeyHome:
			// Go to beginning of line
			row, _ := v.textView.GetScrollOffset()
			v.textView.ScrollTo(row, 0)
			return nil

		case tcell.KeyEnd:
			// Go to end of line
			row, _ := v.textView.GetScrollOffset()
			if row < len(v.lines) {
				lineLen := len(v.lines[row])
				v.textView.ScrollTo(row, lineLen)
			}
			return nil

		case tcell.KeyEscape:
			// Clear search
			if v.searchText != "" {
				v.clearSearch()
				v.updateDisplay()
			}
			return nil
		}

		return event
	})
}

// setupSearchInput configures the search input field
func (v *InspectView) setupSearchInput() {
	v.searchInput.
		SetLabel("Search: ").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetLabelColor(tcell.ColorYellow).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				v.performSearch(v.searchInput.GetText())
				v.isSearchMode = false
			} else if key == tcell.KeyEscape {
				v.isSearchMode = false
				v.searchInput.SetText("")
			}
		})
}

// GetPrimitive returns the tview primitive for this view
func (v *InspectView) GetPrimitive() tview.Primitive {
	if v.isSearchMode {
		// Return a flex container with search input at the bottom
		flex := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(v.textView, 0, 1, false).
			AddItem(v.searchInput, 1, 0, true)
		return flex
	}
	return v.textView
}

// Refresh is not used for inspect view as content is loaded on demand
func (v *InspectView) Refresh() {
	// Not used - content is loaded when switching to this view
}

// GetTitle returns the title of the view
func (v *InspectView) GetTitle() string {
	title := fmt.Sprintf("Inspect: %s (%s)", v.targetName, v.targetType)
	if v.searchText != "" {
		if len(v.searchResults) > 0 {
			title += fmt.Sprintf(" | Search: %s (%d/%d)", v.searchText, v.currentSearchIdx+1, len(v.searchResults))
		} else {
			title += fmt.Sprintf(" | Search: %s (no matches)", v.searchText)
		}
	}
	return title
}

// LoadInspectData loads and displays inspect data for a Docker resource
func (v *InspectView) LoadInspectData(targetName, targetType string) error {
	v.targetName = targetName
	v.targetType = targetType

	slog.Info("Loading inspect data",
		slog.String("target", targetName),
		slog.String("type", targetType))

	// Get inspect data based on type
	var output []byte
	var err error

	switch targetType {
	case "container":
		output, err = docker.ExecuteCaptured("inspect", targetName)
	case "image":
		output, err = docker.ExecuteCaptured("image", "inspect", targetName)
	case "volume":
		output, err = docker.ExecuteCaptured("volume", "inspect", targetName)
	case "network":
		output, err = docker.ExecuteCaptured("network", "inspect", targetName)
	default:
		return fmt.Errorf("unsupported inspect type: %s", targetType)
	}

	if err != nil {
		return fmt.Errorf("failed to inspect %s: %w", targetName, err)
	}

	// Format the JSON data
	formatted, err := v.formatJSON(string(output))
	if err != nil {
		// Fall back to raw JSON if formatting fails
		formatted = string(output)
	}

	v.content = formatted
	v.lines = strings.Split(v.content, "\n")
	v.updateDisplay()

	return nil
}

// formatJSON converts JSON to YAML format for better readability
func (v *InspectView) formatJSON(jsonData string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return "", err
	}

	// Handle array of objects (like docker inspect returns)
	if arr, ok := data.([]interface{}); ok && len(arr) > 0 {
		data = arr[0]
	}

	// Convert to YAML
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

// updateDisplay updates the text view with syntax highlighting
func (v *InspectView) updateDisplay() {
	v.textView.Clear()

	if len(v.lines) == 0 {
		v.textView.SetText("No inspect data available")
		return
	}

	var builder strings.Builder

	for i, line := range v.lines {
		// Add line number
		lineNum := fmt.Sprintf("[gray]%4d[white] ", i+1)
		builder.WriteString(lineNum)

		// Check if this line contains a search match
		isMatch := false
		if v.searchText != "" {
			for _, matchLine := range v.searchResults {
				if matchLine == i {
					isMatch = true
					// Mark current search result
					if len(v.searchResults) > 0 && v.currentSearchIdx < len(v.searchResults) &&
						i == v.searchResults[v.currentSearchIdx] {
						builder.WriteString("[yellow]▶[white] ")
					} else {
						builder.WriteString("  ")
					}
					break
				}
			}
		} else {
			builder.WriteString("  ")
		}

		// Apply syntax highlighting
		highlightedLine := v.applySyntaxHighlighting(line, isMatch)
		builder.WriteString(highlightedLine)
		builder.WriteString("\n")
	}

	v.textView.SetText(builder.String())
}

// applySyntaxHighlighting applies YAML syntax highlighting to a line
func (v *InspectView) applySyntaxHighlighting(line string, isSearchMatch bool) string {
	trimmed := strings.TrimSpace(line)

	// Preserve indentation
	indent := line[:len(line)-len(trimmed)]

	// Apply search highlighting if needed
	// Note: Background colors with : syntax require searchText to be set
	bgColor := ""
	if isSearchMatch && v.searchText != "" {
		bgColor = ":yellow"
	}

	// Check if this line is a YAML key-value pair
	if idx := strings.Index(trimmed, ": "); idx != -1 && !strings.HasPrefix(trimmed, "- ") {
		// Split into key and value parts
		key := trimmed[:idx]
		value := trimmed[idx+2:]

		// Apply colors
		if value == "null" || value == "false" || value == "true" {
			return fmt.Sprintf("%s[blue%s]%s:[white] [red%s]%s[white]", indent, bgColor, key, bgColor, value)
		} else if strings.HasPrefix(value, "\"") || strings.HasPrefix(value, "'") {
			return fmt.Sprintf("%s[blue%s]%s:[white] [green%s]%s[white]", indent, bgColor, key, bgColor, value)
		} else if _, err := json.Number(value).Float64(); err == nil {
			return fmt.Sprintf("%s[blue%s]%s:[white] [yellow%s]%s[white]", indent, bgColor, key, bgColor, value)
		} else {
			return fmt.Sprintf("%s[blue%s]%s:[white] [white%s]%s[white]", indent, bgColor, key, bgColor, value)
		}
	} else if strings.HasPrefix(trimmed, "- ") {
		// YAML list item
		content := trimmed[2:]
		return fmt.Sprintf("%s[white%s]- [green%s]%s[white]", indent, bgColor, bgColor, content)
	} else if trimmed != "" {
		// Other content
		return fmt.Sprintf("%s[white%s]%s[white]", indent, bgColor, trimmed)
	}

	return line
}

// Search functionality
func (v *InspectView) startSearch() {
	v.isSearchMode = true
	v.searchInput.SetText(v.searchText)
	// Focus will be handled by the app when returning the flex primitive
}

func (v *InspectView) performSearch(text string) {
	v.searchText = text
	v.searchResults = nil
	v.currentSearchIdx = 0

	if text == "" {
		v.updateDisplay()
		return
	}

	// Compile regex if needed
	var searchFunc func(string) bool
	if v.searchRegex {
		pattern := text
		if v.searchIgnoreCase {
			pattern = "(?i)" + pattern
		}
		if re, err := regexp.Compile(pattern); err == nil {
			searchFunc = func(line string) bool {
				return re.MatchString(line)
			}
		} else {
			// Invalid regex, fall back to plain text search
			v.searchRegex = false
		}
	}

	// Fall back to plain text search if not regex
	if searchFunc == nil {
		searchStr := text
		if v.searchIgnoreCase {
			searchStr = strings.ToLower(searchStr)
		}
		searchFunc = func(line string) bool {
			lineToSearch := line
			if v.searchIgnoreCase {
				lineToSearch = strings.ToLower(lineToSearch)
			}
			return strings.Contains(lineToSearch, searchStr)
		}
	}

	// Find all matching lines
	for i, line := range v.lines {
		if searchFunc(line) {
			v.searchResults = append(v.searchResults, i)
		}
	}

	// Jump to first result if found
	if len(v.searchResults) > 0 {
		v.currentSearchIdx = 0
		v.jumpToSearchResult(0)
	}

	v.updateDisplay()
}

func (v *InspectView) nextSearchResult() {
	if len(v.searchResults) == 0 {
		return
	}

	v.currentSearchIdx = (v.currentSearchIdx + 1) % len(v.searchResults)
	v.jumpToSearchResult(v.currentSearchIdx)
	v.updateDisplay()
}

func (v *InspectView) prevSearchResult() {
	if len(v.searchResults) == 0 {
		return
	}

	v.currentSearchIdx--
	if v.currentSearchIdx < 0 {
		v.currentSearchIdx = len(v.searchResults) - 1
	}
	v.jumpToSearchResult(v.currentSearchIdx)
	v.updateDisplay()
}

func (v *InspectView) jumpToSearchResult(idx int) {
	if idx >= 0 && idx < len(v.searchResults) {
		targetLine := v.searchResults[idx]
		// Scroll to center the result if possible
		_, _, _, height := v.textView.GetInnerRect()
		scrollTo := targetLine - height/2
		if scrollTo < 0 {
			scrollTo = 0
		}
		v.textView.ScrollTo(scrollTo, 0)
	}
}

func (v *InspectView) clearSearch() {
	v.searchText = ""
	v.searchResults = nil
	v.currentSearchIdx = 0
	v.isSearchMode = false
}

// SetContent sets the inspect content directly (for testing)
func (v *InspectView) SetContent(content, targetName, targetType string) {
	v.content = content
	v.targetName = targetName
	v.targetType = targetType
	v.lines = strings.Split(content, "\n")
	v.updateDisplay()
}

// GetContent returns the current content (for testing)
func (v *InspectView) GetContent() string {
	return v.content
}

// GetSearchResults returns the current search results (for testing)
func (v *InspectView) GetSearchResults() []int {
	return v.searchResults
}

// IsSearchMode returns whether search mode is active (for testing)
func (v *InspectView) IsSearchMode() bool {
	return v.isSearchMode
}

