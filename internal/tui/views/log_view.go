package views

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/tokuhirom/dcv/internal/docker"
)

// LogView displays container logs with search and filter capabilities
type LogView struct {
	docker      *docker.Client
	textView    *tview.TextView
	containerID string
	container   interface{} // Can be Container or other types

	// Log management
	logs         []string
	filteredLogs []string
	mu           sync.RWMutex

	// Streaming
	ctx        context.Context
	cancel     context.CancelFunc
	streaming  bool
	follow     bool
	timestamps bool
	tail       string

	// Search and filter
	searchText       string
	searchRegex      *regexp.Regexp
	searchResults    []int // Line indices of search matches
	currentSearchIdx int
	filterText       string
	filterRegex      *regexp.Regexp
	isFiltered       bool

	// UI state
	autoScroll bool
	wrap       bool
	pages      *tview.Pages
}

// NewLogView creates a new log view
func NewLogView(dockerClient *docker.Client) *LogView {
	v := &LogView{
		docker:     dockerClient,
		textView:   tview.NewTextView(),
		logs:       make([]string, 0),
		autoScroll: true,
		wrap:       true,
		follow:     true,
		timestamps: false,
		tail:       "100",
		pages:      tview.NewPages(),
	}

	v.setupTextView()
	v.setupPages()
	return v
}

// setupTextView configures the text view widget
func (v *LogView) setupTextView() {
	v.textView.
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(v.wrap).
		SetWordWrap(v.wrap).
		SetChangedFunc(func() {
			if v.autoScroll {
				v.textView.ScrollToEnd()
			}
		})

	// Set up key handlers
	v.textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			// Clear search/filter
			if v.searchText != "" || v.isFiltered {
				v.clearSearchAndFilter()
				return nil
			}

		case tcell.KeyPgUp:
			v.autoScroll = false
			row, col := v.textView.GetScrollOffset()
			_, _, _, height := v.textView.GetInnerRect()
			v.textView.ScrollTo(row-height, col)
			return nil

		case tcell.KeyPgDn:
			row, col := v.textView.GetScrollOffset()
			_, _, _, height := v.textView.GetInnerRect()
			v.textView.ScrollTo(row+height, col)
			return nil

		case tcell.KeyHome:
			v.autoScroll = false
			v.textView.ScrollTo(0, 0)
			return nil

		case tcell.KeyEnd:
			v.autoScroll = true
			v.textView.ScrollToEnd()
			return nil
		}

		switch event.Rune() {
		case 'j':
			// Scroll down
			v.autoScroll = false
			row, col := v.textView.GetScrollOffset()
			v.textView.ScrollTo(row+1, col)
			return nil

		case 'k':
			// Scroll up
			v.autoScroll = false
			row, col := v.textView.GetScrollOffset()
			if row > 0 {
				v.textView.ScrollTo(row-1, col)
			}
			return nil

		case 'g':
			// Go to top
			v.autoScroll = false
			v.textView.ScrollTo(0, 0)
			return nil

		case 'G':
			// Go to bottom
			v.autoScroll = true
			v.textView.ScrollToEnd()
			return nil

		case '/':
			// Start search
			v.showSearchInput()
			return nil

		case 'f', 'F':
			// Start filter
			v.showFilterInput()
			return nil

		case 'n':
			// Next search result
			v.nextSearchResult()
			return nil

		case 'N':
			// Previous search result
			v.prevSearchResult()
			return nil

		case 'w':
			// Toggle wrap
			v.wrap = !v.wrap
			v.textView.SetWrap(v.wrap).SetWordWrap(v.wrap)
			return nil

		case 't':
			// Toggle timestamps
			v.timestamps = !v.timestamps
			v.reloadLogs()
			return nil

		case 'a':
			// Toggle auto-scroll
			v.autoScroll = !v.autoScroll
			if v.autoScroll {
				v.textView.ScrollToEnd()
			}
			return nil

		case 'p':
			// Toggle pause/resume streaming
			if v.follow {
				v.pauseStreaming()
			} else {
				v.resumeStreaming()
			}
			return nil

		case 'c':
			// Clear logs
			v.clearLogs()
			return nil

		case 'r':
			// Refresh logs
			v.reloadLogs()
			return nil
		}

		return event
	})
}

// setupPages sets up the pages for search/filter input
func (v *LogView) setupPages() {
	v.pages.AddPage("logs", v.textView, true, true)
}

// GetPrimitive returns the tview primitive for this view
func (v *LogView) GetPrimitive() tview.Primitive {
	return v.pages
}

// Refresh refreshes the log view
func (v *LogView) Refresh() {
	if v.containerID != "" && !v.streaming {
		v.startStreaming()
	}
}

// GetTitle returns the title of the view
func (v *LogView) GetTitle() string {
	title := "Container Logs"
	if v.containerID != "" {
		title = fmt.Sprintf("Logs: %s", v.containerID[:12])
	}

	// Add status indicators
	var status []string
	if v.follow {
		status = append(status, "Following")
	} else {
		status = append(status, "Paused")
	}

	if v.autoScroll {
		status = append(status, "Auto-scroll")
	}

	if v.wrap {
		status = append(status, "Wrap")
	}

	if v.timestamps {
		status = append(status, "Timestamps")
	}

	if v.searchText != "" {
		status = append(status, fmt.Sprintf("Search: %s", v.searchText))
	}

	if v.isFiltered {
		status = append(status, fmt.Sprintf("Filter: %s", v.filterText))
	}

	if len(status) > 0 {
		title += " [" + strings.Join(status, " | ") + "]"
	}

	return title
}

// SetContainer sets the container to view logs for
func (v *LogView) SetContainer(containerID string, container interface{}) {
	v.mu.Lock()
	v.containerID = containerID
	v.container = container
	v.mu.Unlock()

	// Stop any existing streaming
	v.stopStreaming()

	// Clear logs and start new stream
	v.clearLogs()
	v.startStreaming()
}

// startStreaming starts streaming logs from the container
func (v *LogView) startStreaming() {
	if v.containerID == "" {
		return
	}

	v.mu.Lock()
	if v.streaming {
		v.mu.Unlock()
		return
	}
	v.streaming = true
	v.mu.Unlock()

	// Create context for cancellation
	v.ctx, v.cancel = context.WithCancel(context.Background())

	// Start streaming in background
	go v.streamLogs()
}

// stopStreaming stops streaming logs
func (v *LogView) stopStreaming() {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.cancel != nil {
		v.cancel()
		v.cancel = nil
	}
	v.streaming = false
}

// streamLogs streams logs from the container
func (v *LogView) streamLogs() {
	args := []string{"logs"}

	if v.follow {
		args = append(args, "-f")
	}

	if v.timestamps {
		args = append(args, "-t")
	}

	if v.tail != "" && v.tail != "all" {
		args = append(args, "--tail", v.tail)
	}

	args = append(args, v.containerID)

	// Execute docker logs command
	reader, err := docker.ExecuteStreamingCommand(v.ctx, args...)
	if err != nil {
		slog.Error("Failed to stream logs", slog.Any("error", err))
		v.addLog(fmt.Sprintf("[red]Error: %v[white]", err))
		v.mu.Lock()
		v.streaming = false
		v.mu.Unlock()
		return
	}

	// Read logs line by line
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		select {
		case <-v.ctx.Done():
			return
		default:
			line := scanner.Text()
			v.addLog(line)
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		slog.Error("Error reading logs", slog.Any("error", err))
		v.addLog(fmt.Sprintf("[red]Error reading logs: %v[white]", err))
	}

	v.mu.Lock()
	v.streaming = false
	v.mu.Unlock()
}

// addLog adds a log line
func (v *LogView) addLog(line string) {
	v.mu.Lock()
	v.logs = append(v.logs, line)

	// Limit log buffer size (keep last 10000 lines)
	if len(v.logs) > 10000 {
		v.logs = v.logs[len(v.logs)-10000:]
	}
	v.mu.Unlock()

	// Update display
	QueueUpdateDraw(func() {
		v.updateDisplay()
	})
}

// updateDisplay updates the text view with current logs
func (v *LogView) updateDisplay() {
	v.mu.RLock()
	logs := v.logs
	if v.isFiltered && v.filterRegex != nil {
		logs = v.filteredLogs
	}
	v.mu.RUnlock()

	// Build display text with search highlighting
	var builder strings.Builder
	for i, line := range logs {
		displayLine := line

		// Highlight search matches
		if v.searchText != "" && v.searchRegex != nil {
			if v.searchRegex.MatchString(line) {
				// Simple highlight by adding color tags
				displayLine = v.searchRegex.ReplaceAllString(line, "[yellow::b]$0[white::-]")
			}
		}

		// Add line number if searching
		if v.searchText != "" {
			builder.WriteString(fmt.Sprintf("[gray]%5d[white] ", i+1))
		}

		builder.WriteString(displayLine)
		builder.WriteString("\n")
	}

	v.textView.SetText(builder.String())
}

// showSearchInput shows the search input dialog
func (v *LogView) showSearchInput() {
	inputField := tview.NewInputField().
		SetLabel("Search: ").
		SetText(v.searchText).
		SetFieldWidth(50)

	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			v.performSearch(inputField.GetText())
		}
		v.pages.SwitchToPage("logs")
	})

	// Create modal
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(inputField, 1, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)

	v.pages.AddPage("search", modal, true, true)
}

// performSearch performs a search in the logs
func (v *LogView) performSearch(text string) {
	v.searchText = text
	v.searchResults = nil
	v.currentSearchIdx = 0

	if text == "" {
		v.searchRegex = nil
		v.updateDisplay()
		return
	}

	// Compile regex
	var err error
	v.searchRegex, err = regexp.Compile("(?i)" + regexp.QuoteMeta(text))
	if err != nil {
		slog.Error("Invalid search pattern", slog.Any("error", err))
		return
	}

	// Find all matching lines
	v.mu.RLock()
	logs := v.logs
	if v.isFiltered {
		logs = v.filteredLogs
	}
	v.mu.RUnlock()

	for i, line := range logs {
		if v.searchRegex.MatchString(line) {
			v.searchResults = append(v.searchResults, i)
		}
	}

	// Jump to first result
	if len(v.searchResults) > 0 {
		v.jumpToSearchResult(0)
	}

	v.updateDisplay()
}

// nextSearchResult jumps to the next search result
func (v *LogView) nextSearchResult() {
	if len(v.searchResults) == 0 {
		return
	}

	v.currentSearchIdx = (v.currentSearchIdx + 1) % len(v.searchResults)
	v.jumpToSearchResult(v.currentSearchIdx)
}

// prevSearchResult jumps to the previous search result
func (v *LogView) prevSearchResult() {
	if len(v.searchResults) == 0 {
		return
	}

	v.currentSearchIdx--
	if v.currentSearchIdx < 0 {
		v.currentSearchIdx = len(v.searchResults) - 1
	}
	v.jumpToSearchResult(v.currentSearchIdx)
}

// jumpToSearchResult scrolls to a specific search result
func (v *LogView) jumpToSearchResult(idx int) {
	if idx >= 0 && idx < len(v.searchResults) {
		lineNum := v.searchResults[idx]
		// Calculate approximate position
		v.autoScroll = false
		v.textView.ScrollTo(lineNum, 0)
	}
}

// showFilterInput shows the filter input dialog
func (v *LogView) showFilterInput() {
	inputField := tview.NewInputField().
		SetLabel("Filter: ").
		SetText(v.filterText).
		SetFieldWidth(50)

	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			v.applyFilter(inputField.GetText())
		}
		v.pages.SwitchToPage("logs")
	})

	// Create modal
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(inputField, 1, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)

	v.pages.AddPage("filter", modal, true, true)
}

// applyFilter applies a filter to the logs
func (v *LogView) applyFilter(text string) {
	v.filterText = text

	if text == "" {
		v.isFiltered = false
		v.filterRegex = nil
		v.filteredLogs = nil
		v.updateDisplay()
		return
	}

	// Compile regex
	var err error
	v.filterRegex, err = regexp.Compile("(?i)" + regexp.QuoteMeta(text))
	if err != nil {
		slog.Error("Invalid filter pattern", slog.Any("error", err))
		return
	}

	// Filter logs
	v.mu.Lock()
	v.filteredLogs = nil
	for _, line := range v.logs {
		if v.filterRegex.MatchString(line) {
			v.filteredLogs = append(v.filteredLogs, line)
		}
	}
	v.isFiltered = true
	v.mu.Unlock()

	v.updateDisplay()
}

// clearSearchAndFilter clears search and filter
func (v *LogView) clearSearchAndFilter() {
	v.searchText = ""
	v.searchRegex = nil
	v.searchResults = nil
	v.currentSearchIdx = 0
	v.filterText = ""
	v.filterRegex = nil
	v.isFiltered = false
	v.filteredLogs = nil
	v.updateDisplay()
}

// clearLogs clears all logs
func (v *LogView) clearLogs() {
	v.mu.Lock()
	v.logs = nil
	v.filteredLogs = nil
	v.mu.Unlock()
	v.updateDisplay()
}

// reloadLogs reloads logs from the container
func (v *LogView) reloadLogs() {
	v.stopStreaming()
	v.clearLogs()
	v.startStreaming()
}

// pauseStreaming pauses log streaming
func (v *LogView) pauseStreaming() {
	v.follow = false
	v.stopStreaming()
}

// resumeStreaming resumes log streaming
func (v *LogView) resumeStreaming() {
	v.follow = true
	v.startStreaming()
}

// Stop stops the log view (cleanup)
func (v *LogView) Stop() {
	v.stopStreaming()
}

// GetContainerID returns the current container ID
func (v *LogView) GetContainerID() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.containerID
}

// SetTailLines sets the number of lines to tail
func (v *LogView) SetTailLines(lines string) {
	v.tail = lines
	if v.streaming {
		v.reloadLogs()
	}
}

// IsStreaming returns whether logs are currently streaming
func (v *LogView) IsStreaming() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.streaming
}
