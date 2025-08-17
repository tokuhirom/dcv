package views

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// FileBrowserView displays files and directories inside a container
type FileBrowserView struct {
	docker            *docker.Client
	table             *tview.Table
	containerID       string
	containerName     string
	currentPath       string
	files             []models.ContainerFile
	pathHistory       []string
	isLoading         bool
	mu                sync.RWMutex
	switchToContentFn func(containerID, path string, file models.ContainerFile)
}

// NewFileBrowserView creates a new file browser view
func NewFileBrowserView(dockerClient *docker.Client) *FileBrowserView {
	v := &FileBrowserView{
		docker:      dockerClient,
		table:       tview.NewTable(),
		currentPath: "/",
		pathHistory: []string{"/"},
	}

	v.setupTable()
	v.setupKeyHandlers()

	return v
}

// setupTable configures the table widget
func (v *FileBrowserView) setupTable() {
	v.table.SetBorders(false).
		SetSelectable(true, false).
		SetSeparator(' ').
		SetFixed(1, 0)

	// Set header style
	v.table.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.ColorDarkCyan).
		Foreground(tcell.ColorWhite))
}

// setupKeyHandlers sets up keyboard shortcuts for the view
func (v *FileBrowserView) setupKeyHandlers() {
	v.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := v.table.GetSelection()

		switch event.Rune() {
		case 'j':
			// Move down (vim style)
			if row < v.table.GetRowCount()-1 {
				v.table.Select(row+1, 0)
			}
			return nil

		case 'k':
			// Move up (vim style)
			if row > 1 { // Skip header row
				v.table.Select(row-1, 0)
			}
			return nil

		case 'g':
			// Go to top (vim style)
			v.table.Select(1, 0)
			return nil

		case 'G':
			// Go to bottom (vim style)
			rowCount := v.table.GetRowCount()
			if rowCount > 1 {
				v.table.Select(rowCount-1, 0)
			}
			return nil

		case 'r':
			// Refresh current directory
			v.loadFiles()
			return nil

		case 'h', '-':
			// Go back to parent directory
			v.navigateUp()
			return nil

		case 'H':
			// Go to root directory
			v.navigateToRoot()
			return nil

		case '~':
			// Go to home directory
			v.navigateTo("/root")
			return nil
		}

		switch event.Key() {
		case tcell.KeyEnter:
			// Enter directory or view file
			v.handleEnter()
			return nil

		case tcell.KeyBackspace, tcell.KeyBackspace2:
			// Go back to parent directory
			v.navigateUp()
			return nil

		case tcell.KeyEscape:
			// Go back in history
			v.navigateBack()
			return nil

		case tcell.KeyCtrlR:
			// Force refresh
			v.loadFiles()
			return nil
		}

		return event
	})
}

// GetPrimitive returns the tview primitive for this view
func (v *FileBrowserView) GetPrimitive() tview.Primitive {
	return v.table
}

// Refresh refreshes the view's data
func (v *FileBrowserView) Refresh() {
	go v.loadFiles()
}

// GetTitle returns the title of the view
func (v *FileBrowserView) GetTitle() string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.containerName != "" {
		return fmt.Sprintf("File Browser: %s:%s", v.containerName, v.currentPath)
	}
	return fmt.Sprintf("File Browser: %s", v.currentPath)
}

// SetContainer sets the container to browse files in
func (v *FileBrowserView) SetContainer(containerID, containerName string) {
	v.mu.Lock()
	v.containerID = containerID
	v.containerName = containerName
	v.currentPath = "/"
	v.pathHistory = []string{"/"}
	v.mu.Unlock()

	// Load files asynchronously to avoid blocking the UI
	go v.loadFiles()
}

// SetSwitchToContentViewCallback sets the callback for switching to content view
func (v *FileBrowserView) SetSwitchToContentViewCallback(fn func(containerID, path string, file models.ContainerFile)) {
	v.switchToContentFn = fn
}

// loadFiles loads the file list for the current path
func (v *FileBrowserView) loadFiles() {
	v.mu.RLock()
	containerID := v.containerID
	currentPath := v.currentPath
	v.mu.RUnlock()

	if containerID == "" {
		return
	}

	// Set loading state
	v.mu.Lock()
	v.isLoading = true
	v.mu.Unlock()

	// Show loading message
	QueueUpdateDraw(func() {
		v.showLoadingMessage()
	})

	slog.Info("Loading files",
		slog.String("container", containerID),
		slog.String("path", currentPath))

	files, err := v.docker.ListContainerFiles(containerID, currentPath)

	v.mu.Lock()
	v.isLoading = false
	if err != nil {
		slog.Error("Failed to load files", slog.Any("error", err))
		v.files = []models.ContainerFile{}
	} else {
		v.files = files
	}
	v.mu.Unlock()

	// Update table in UI thread
	QueueUpdateDraw(func() {
		v.updateTable()
	})
}

// showLoadingMessage displays a loading message in the table
func (v *FileBrowserView) showLoadingMessage() {
	v.table.Clear()
	cell := tview.NewTableCell("Loading files...").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignCenter)
	v.table.SetCell(0, 0, cell)
}

// updateTable updates the table with file data
func (v *FileBrowserView) updateTable() {
	v.table.Clear()

	v.mu.RLock()
	files := v.files
	currentPath := v.currentPath
	v.mu.RUnlock()

	// Set headers
	headers := []string{"TYPE", "PERMISSIONS", "SIZE", "MODIFIED", "NAME"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false)
		v.table.SetCell(0, col, cell)
	}

	// Add parent directory entry if not at root
	row := 1
	if currentPath != "/" {
		typeCell := tview.NewTableCell("DIR").
			SetTextColor(tcell.ColorBlue)
		v.table.SetCell(row, 0, typeCell)

		permCell := tview.NewTableCell("").
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row, 1, permCell)

		sizeCell := tview.NewTableCell("").
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row, 2, sizeCell)

		modCell := tview.NewTableCell("").
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row, 3, modCell)

		nameCell := tview.NewTableCell("..").
			SetTextColor(tcell.ColorBlue).
			SetAttributes(tcell.AttrBold)
		v.table.SetCell(row, 4, nameCell)

		row++
	}

	// Add file rows
	for _, file := range files {
		// Type
		typeStr := "FILE"
		typeColor := tcell.ColorWhite
		if file.IsDir {
			typeStr = "DIR"
			typeColor = tcell.ColorBlue
		} else if file.LinkTarget != "" {
			typeStr = "LINK"
			typeColor = tcell.ColorDarkCyan
		}
		typeCell := tview.NewTableCell(typeStr).
			SetTextColor(typeColor)
		v.table.SetCell(row, 0, typeCell)

		// Permissions
		permCell := tview.NewTableCell(file.Permissions).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row, 1, permCell)

		// Size
		sizeStr := formatFileSize(file.Size)
		if file.IsDir {
			sizeStr = "-"
		}
		sizeCell := tview.NewTableCell(sizeStr).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row, 2, sizeCell)

		// Modified time
		modCell := tview.NewTableCell(file.ModTime.Format("2006-01-02 15:04")).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row, 3, modCell)

		// Name
		name := file.Name
		nameColor := tcell.ColorWhite
		nameAttrs := tcell.AttrNone
		if file.IsDir {
			name = name + "/"
			nameColor = tcell.ColorBlue
			nameAttrs = tcell.AttrBold
		} else if file.LinkTarget != "" {
			name = fmt.Sprintf("%s -> %s", name, file.LinkTarget)
			nameColor = tcell.ColorDarkCyan
		} else if isExecutable(file.Permissions) {
			nameColor = tcell.ColorGreen
			nameAttrs = tcell.AttrBold
		}
		nameCell := tview.NewTableCell(name).
			SetTextColor(nameColor).
			SetAttributes(nameAttrs)
		v.table.SetCell(row, 4, nameCell)

		row++
	}

	// Select first row if available
	if v.table.GetRowCount() > 1 {
		v.table.Select(1, 0)
	}
}

// handleEnter handles the Enter key press
func (v *FileBrowserView) handleEnter() {
	row, _ := v.table.GetSelection()
	if row < 1 {
		return
	}

	v.mu.RLock()
	currentPath := v.currentPath
	files := v.files
	containerID := v.containerID
	v.mu.RUnlock()

	// Check if it's the parent directory entry
	if currentPath != "/" && row == 1 {
		v.navigateUp()
		return
	}

	// Adjust index for parent directory entry
	fileIndex := row - 1
	if currentPath != "/" {
		fileIndex = row - 2
	}

	if fileIndex < 0 || fileIndex >= len(files) {
		return
	}

	file := files[fileIndex]

	if file.IsDir {
		// Navigate into directory
		newPath := filepath.Join(currentPath, file.Name)
		v.navigateTo(newPath)
	} else if v.switchToContentFn != nil {
		// View file content
		filePath := filepath.Join(currentPath, file.Name)
		v.switchToContentFn(containerID, filePath, file)
	} else {
		slog.Info("View file content",
			slog.String("container", containerID),
			slog.String("file", file.Name))
	}
}

// navigateTo navigates to a specific path
func (v *FileBrowserView) navigateTo(path string) {
	v.mu.Lock()
	v.currentPath = path
	v.pathHistory = append(v.pathHistory, path)
	v.mu.Unlock()

	go v.loadFiles()
}

// navigateUp navigates to the parent directory
func (v *FileBrowserView) navigateUp() {
	v.mu.RLock()
	currentPath := v.currentPath
	v.mu.RUnlock()

	if currentPath == "/" {
		return
	}

	parentPath := filepath.Dir(currentPath)
	v.navigateTo(parentPath)
}

// navigateBack goes back in navigation history
func (v *FileBrowserView) navigateBack() {
	v.mu.Lock()
	if len(v.pathHistory) <= 1 {
		v.mu.Unlock()
		return
	}

	// Remove current path from history
	v.pathHistory = v.pathHistory[:len(v.pathHistory)-1]
	// Set current path to the previous one
	v.currentPath = v.pathHistory[len(v.pathHistory)-1]
	v.mu.Unlock()

	go v.loadFiles()
}

// navigateToRoot navigates to the root directory
func (v *FileBrowserView) navigateToRoot() {
	v.navigateTo("/")
}

// GetSelectedFile returns the currently selected file
func (v *FileBrowserView) GetSelectedFile() *models.ContainerFile {
	row, _ := v.table.GetSelection()
	if row < 1 {
		return nil
	}

	v.mu.RLock()
	defer v.mu.RUnlock()

	// Check if it's the parent directory entry
	if v.currentPath != "/" && row == 1 {
		return nil
	}

	// Adjust index for parent directory entry
	fileIndex := row - 1
	if v.currentPath != "/" {
		fileIndex = row - 2
	}

	if fileIndex < 0 || fileIndex >= len(v.files) {
		return nil
	}

	return &v.files[fileIndex]
}

// GetContainerName returns the current container name
func (v *FileBrowserView) GetContainerName() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.containerName
}

// GetContainerID returns the current container ID
func (v *FileBrowserView) GetContainerID() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.containerID
}

// formatFileSize formats file size in human-readable format
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// isExecutable checks if a file is executable based on its permissions
func isExecutable(permissions string) bool {
	if len(permissions) < 10 {
		return false
	}
	// Check for any execute permission (user, group, or other)
	return permissions[3] == 'x' || permissions[6] == 'x' || permissions[9] == 'x'
}
