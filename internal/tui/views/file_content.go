package views

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// FileContentView displays the content of a file from a container
type FileContentView struct {
	docker        *docker.Client
	textView      *tview.TextView
	containerID   string
	containerName string
	filePath      string
	fileInfo      models.ContainerFile
	content       string
	lineNumbers   bool
	wrap          bool
	mu            sync.RWMutex
}

// NewFileContentView creates a new file content view
func NewFileContentView(dockerClient *docker.Client) *FileContentView {
	v := &FileContentView{
		docker:      dockerClient,
		textView:    tview.NewTextView(),
		lineNumbers: true,
		wrap:        false,
	}

	v.setupTextView()
	v.setupKeyHandlers()

	return v
}

// setupTextView configures the text view widget
func (v *FileContentView) setupTextView() {
	v.textView.SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(v.wrap).
		SetWordWrap(false)
}

// setupKeyHandlers sets up keyboard shortcuts for the view
func (v *FileContentView) setupKeyHandlers() {
	v.textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
			// Scroll left (vim style)
			row, col := v.textView.GetScrollOffset()
			if col > 0 {
				v.textView.ScrollTo(row, col-1)
			}
			return nil

		case 'l':
			// Scroll right (vim style)
			row, col := v.textView.GetScrollOffset()
			v.textView.ScrollTo(row, col+1)
			return nil

		case 'n':
			// Toggle line numbers
			v.toggleLineNumbers()
			return nil

		case 'w':
			// Toggle word wrap
			v.toggleWrap()
			return nil

		case 'r':
			// Refresh content
			v.loadContent()
			return nil

		case 'd':
			// Download file (placeholder for future implementation)
			slog.Info("Download file requested",
				slog.String("container", v.containerID),
				slog.String("path", v.filePath))
			return nil
		}

		switch event.Key() {
		case tcell.KeyCtrlD:
			// Page down
			_, _, _, height := v.textView.GetInnerRect()
			row, col := v.textView.GetScrollOffset()
			v.textView.ScrollTo(row+height/2, col)
			return nil

		case tcell.KeyCtrlU:
			// Page up
			_, _, _, height := v.textView.GetInnerRect()
			row, col := v.textView.GetScrollOffset()
			newRow := row - height/2
			if newRow < 0 {
				newRow = 0
			}
			v.textView.ScrollTo(newRow, col)
			return nil

		case tcell.KeyCtrlF:
			// Page down (full page)
			_, _, _, height := v.textView.GetInnerRect()
			row, col := v.textView.GetScrollOffset()
			v.textView.ScrollTo(row+height-1, col)
			return nil

		case tcell.KeyCtrlB:
			// Page up (full page)
			_, _, _, height := v.textView.GetInnerRect()
			row, col := v.textView.GetScrollOffset()
			newRow := row - height + 1
			if newRow < 0 {
				newRow = 0
			}
			v.textView.ScrollTo(newRow, col)
			return nil

		case tcell.KeyCtrlR:
			// Force refresh
			v.loadContent()
			return nil

		case tcell.KeyHome:
			// Go to beginning of line
			row, _ := v.textView.GetScrollOffset()
			v.textView.ScrollTo(row, 0)
			return nil

		case tcell.KeyEnd:
			// Go to end of line
			row, _ := v.textView.GetScrollOffset()
			v.textView.ScrollTo(row, 999999)
			return nil
		}

		return event
	})
}

// GetPrimitive returns the tview primitive for this view
func (v *FileContentView) GetPrimitive() tview.Primitive {
	return v.textView
}

// Refresh refreshes the view's data
func (v *FileContentView) Refresh() {
	go v.loadContent()
}

// GetTitle returns the title of the view
func (v *FileContentView) GetTitle() string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	title := fmt.Sprintf("File: %s", v.filePath)
	if v.containerName != "" {
		title = fmt.Sprintf("File: %s:%s", v.containerName, v.filePath)
	}

	// Add file info
	if v.fileInfo.Size > 0 {
		title += fmt.Sprintf(" [%s]", formatFileSize(v.fileInfo.Size))
	}

	// Add view options
	options := []string{}
	if v.lineNumbers {
		options = append(options, "Line numbers")
	}
	if v.wrap {
		options = append(options, "Wrap")
	}

	if len(options) > 0 {
		title += fmt.Sprintf(" [%s]", strings.Join(options, " | "))
	}

	return title
}

// SetFile sets the file to display
func (v *FileContentView) SetFile(containerID, containerName, filePath string, fileInfo models.ContainerFile) {
	v.mu.Lock()
	v.containerID = containerID
	v.containerName = containerName
	v.filePath = filePath
	v.fileInfo = fileInfo
	v.mu.Unlock()

	// Load content asynchronously to avoid blocking the UI
	go v.loadContent()
}

// loadContent loads the file content from the container
func (v *FileContentView) loadContent() {
	v.mu.RLock()
	containerID := v.containerID
	filePath := v.filePath
	v.mu.RUnlock()

	if containerID == "" || filePath == "" {
		return
	}

	// Show loading message
	QueueUpdateDraw(func() {
		v.textView.SetText("[yellow]Loading file content...[-]")
	})

	slog.Info("Loading file content",
		slog.String("container", containerID),
		slog.String("path", filePath))

	content, err := v.docker.GetFileContent(containerID, filePath)
	if err != nil {
		slog.Error("Failed to load file content", slog.Any("error", err))
		v.mu.Lock()
		v.content = fmt.Sprintf("[red]Error loading file: %v[-]", err)
		v.mu.Unlock()
	} else {
		v.mu.Lock()
		v.content = content
		v.mu.Unlock()
	}

	// Update text view in UI thread
	QueueUpdateDraw(func() {
		v.updateContent()
	})
}

// updateContent updates the text view with the file content
func (v *FileContentView) updateContent() {
	v.mu.RLock()
	content := v.content
	lineNumbers := v.lineNumbers
	v.mu.RUnlock()

	if lineNumbers && !strings.HasPrefix(content, "[red]Error") {
		// Add line numbers
		lines := strings.Split(content, "\n")
		var numberedContent strings.Builder
		lineNumWidth := len(fmt.Sprintf("%d", len(lines)))

		for i, line := range lines {
			lineNum := i + 1
			numberedContent.WriteString(fmt.Sprintf("[gray]%*d[-] %s\n", lineNumWidth, lineNum, line))
		}

		v.textView.SetText(numberedContent.String())
	} else {
		v.textView.SetText(content)
	}

	// Scroll to top when new content is loaded
	v.textView.ScrollTo(0, 0)
}

// toggleLineNumbers toggles line number display
func (v *FileContentView) toggleLineNumbers() {
	v.mu.Lock()
	v.lineNumbers = !v.lineNumbers
	v.mu.Unlock()

	QueueUpdateDraw(func() {
		v.updateContent()
	})
}

// toggleWrap toggles text wrapping
func (v *FileContentView) toggleWrap() {
	v.mu.Lock()
	v.wrap = !v.wrap
	v.mu.Unlock()

	QueueUpdateDraw(func() {
		v.textView.SetWrap(v.wrap)
	})
}

// GetContent returns the current file content
func (v *FileContentView) GetContent() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.content
}

// GetFilePath returns the current file path
func (v *FileContentView) GetFilePath() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.filePath
}
