package views

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// ProjectListView displays Docker Compose projects
type ProjectListView struct {
	docker   *docker.Client
	table    *tview.Table
	projects []models.ComposeProject

	// Callback for when a project is selected
	onProjectSelected func(project models.ComposeProject)
}

// NewProjectListView creates a new project list view
func NewProjectListView(dockerClient *docker.Client) *ProjectListView {
	v := &ProjectListView{
		docker: dockerClient,
		table:  tview.NewTable(),
	}

	v.setupTable()
	v.setupKeyHandlers()

	return v
}

// setupTable configures the table widget
func (v *ProjectListView) setupTable() {
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
func (v *ProjectListView) setupKeyHandlers() {
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

		case 's':
			// Stop project
			if row > 0 && row <= len(v.projects) {
				project := v.projects[row-1]
				go v.stopProject(project)
			}
			return nil

		case 'S':
			// Start project
			if row > 0 && row <= len(v.projects) {
				project := v.projects[row-1]
				go v.startProject(project)
			}
			return nil

		case 'd':
			// Down (stop) project
			if row > 0 && row <= len(v.projects) {
				project := v.projects[row-1]
				go v.downProject(project)
			}
			return nil

		case 'D':
			// Remove project (down and remove volumes)
			if row > 0 && row <= len(v.projects) {
				project := v.projects[row-1]
				go v.removeProject(project)
			}
			return nil

		case 'u':
			// Up project
			if row > 0 && row <= len(v.projects) {
				project := v.projects[row-1]
				go v.upProject(project)
			}
			return nil

		case 'U':
			// Up project in detached mode
			if row > 0 && row <= len(v.projects) {
				project := v.projects[row-1]
				go v.upProjectDetached(project)
			}
			return nil

		case 'r':
			// Restart project
			if row > 0 && row <= len(v.projects) {
				project := v.projects[row-1]
				go v.restartProject(project)
			}
			return nil

		case 'l':
			// View logs
			if row > 0 && row <= len(v.projects) {
				project := v.projects[row-1]
				// TODO: Switch to log view for the project
				slog.Info("View project logs", slog.String("project", project.Name))
			}
			return nil

		case 'p':
			// Pull images for project
			if row > 0 && row <= len(v.projects) {
				project := v.projects[row-1]
				go v.pullProject(project)
			}
			return nil

		case 'b':
			// Build project
			if row > 0 && row <= len(v.projects) {
				project := v.projects[row-1]
				go v.buildProject(project)
			}
			return nil

		case 'R':
			// Refresh list
			v.Refresh()
			return nil

		case '/':
			// Search projects
			// TODO: Implement search functionality
			slog.Info("Search projects")
			return nil
		}

		switch event.Key() {
		case tcell.KeyEnter:
			// Select project and switch to compose process list
			if row > 0 && row <= len(v.projects) {
				project := v.projects[row-1]
				if v.onProjectSelected != nil {
					v.onProjectSelected(project)
				} else {
					slog.Info("Project selected", slog.String("project", project.Name))
				}
			}
			return nil

		case tcell.KeyCtrlR:
			// Force refresh
			v.Refresh()
			return nil
		}

		return event
	})
}

// SetOnProjectSelected sets the callback for when a project is selected
func (v *ProjectListView) SetOnProjectSelected(callback func(project models.ComposeProject)) {
	v.onProjectSelected = callback
}

// GetPrimitive returns the tview primitive for this view
func (v *ProjectListView) GetPrimitive() tview.Primitive {
	return v.table
}

// Refresh refreshes the project list
func (v *ProjectListView) Refresh() {
	go v.loadProjects()
}

// GetTitle returns the title of the view
func (v *ProjectListView) GetTitle() string {
	return "Docker Compose Projects"
}

// loadProjects loads the project list from Docker Compose
func (v *ProjectListView) loadProjects() {
	slog.Info("Loading compose projects")

	projects, err := v.docker.ListComposeProjects()
	if err != nil {
		slog.Error("Failed to load compose projects", slog.Any("error", err))
		return
	}

	v.projects = projects

	// Update table in UI thread
	QueueUpdateDraw(func() {
		v.updateTable()
	})
}

// updateTable updates the table with project data
func (v *ProjectListView) updateTable() {
	v.table.Clear()

	// Set headers
	headers := []string{"NAME", "STATUS", "CONFIG FILES"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false)
		v.table.SetCell(0, col, cell)
	}

	// Add project rows
	for row, project := range v.projects {
		// Project name
		nameCell := tview.NewTableCell(project.Name).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 0, nameCell)

		// Status with color
		status := project.Status
		statusColor := tcell.ColorRed
		if strings.Contains(strings.ToLower(status), "running") {
			statusColor = tcell.ColorGreen
		} else if strings.Contains(strings.ToLower(status), "exited") {
			statusColor = tcell.ColorYellow
		}
		statusCell := tview.NewTableCell(status).
			SetTextColor(statusColor)
		v.table.SetCell(row+1, 1, statusCell)

		// Config files
		configCell := tview.NewTableCell(project.ConfigFiles).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 2, configCell)
	}

	// Select first row if available
	if len(v.projects) > 0 {
		v.table.Select(1, 0)
	}
}

// Project operations
func (v *ProjectListView) stopProject(project models.ComposeProject) {
	slog.Info("Stopping project", slog.String("project", project.Name))

	_, err := docker.ExecuteCaptured("compose", "-p", project.Name, "stop")
	if err != nil {
		slog.Error("Failed to stop project", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ProjectListView) startProject(project models.ComposeProject) {
	slog.Info("Starting project", slog.String("project", project.Name))

	_, err := docker.ExecuteCaptured("compose", "-p", project.Name, "start")
	if err != nil {
		slog.Error("Failed to start project", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ProjectListView) downProject(project models.ComposeProject) {
	slog.Info("Bringing down project", slog.String("project", project.Name))

	_, err := docker.ExecuteCaptured("compose", "-p", project.Name, "down")
	if err != nil {
		slog.Error("Failed to bring down project", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ProjectListView) removeProject(project models.ComposeProject) {
	slog.Info("Removing project with volumes", slog.String("project", project.Name))

	_, err := docker.ExecuteCaptured("compose", "-p", project.Name, "down", "-v")
	if err != nil {
		slog.Error("Failed to remove project", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ProjectListView) upProject(project models.ComposeProject) {
	slog.Info("Bringing up project", slog.String("project", project.Name))

	// Note: This will run in foreground, which might not be ideal for TUI
	// Consider using upProjectDetached instead
	_, err := docker.ExecuteCaptured("compose", "-p", project.Name, "up")
	if err != nil {
		slog.Error("Failed to bring up project", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ProjectListView) upProjectDetached(project models.ComposeProject) {
	slog.Info("Bringing up project in detached mode", slog.String("project", project.Name))

	_, err := docker.ExecuteCaptured("compose", "-p", project.Name, "up", "-d")
	if err != nil {
		slog.Error("Failed to bring up project", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ProjectListView) restartProject(project models.ComposeProject) {
	slog.Info("Restarting project", slog.String("project", project.Name))

	_, err := docker.ExecuteCaptured("compose", "-p", project.Name, "restart")
	if err != nil {
		slog.Error("Failed to restart project", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ProjectListView) pullProject(project models.ComposeProject) {
	slog.Info("Pulling images for project", slog.String("project", project.Name))

	_, err := docker.ExecuteCaptured("compose", "-p", project.Name, "pull")
	if err != nil {
		slog.Error("Failed to pull images for project", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ProjectListView) buildProject(project models.ComposeProject) {
	slog.Info("Building project", slog.String("project", project.Name))

	_, err := docker.ExecuteCaptured("compose", "-p", project.Name, "build")
	if err != nil {
		slog.Error("Failed to build project", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

// GetSelectedProject returns the currently selected project
func (v *ProjectListView) GetSelectedProject() *models.ComposeProject {
	row, _ := v.table.GetSelection()
	if row > 0 && row <= len(v.projects) {
		return &v.projects[row-1]
	}
	return nil
}

// SearchProjects searches for projects in the list
func (v *ProjectListView) SearchProjects(query string) {
	if query == "" {
		// Reset to show all loaded projects
		v.updateTable()
		return
	}

	// Filter projects based on query
	var filteredProjects []models.ComposeProject
	lowerQuery := strings.ToLower(query)

	for _, project := range v.projects {
		if strings.Contains(strings.ToLower(project.Name), lowerQuery) ||
			strings.Contains(strings.ToLower(project.Status), lowerQuery) ||
			strings.Contains(strings.ToLower(project.ConfigFiles), lowerQuery) {
			filteredProjects = append(filteredProjects, project)
		}
	}

	// Temporarily replace projects for display
	originalProjects := v.projects
	v.projects = filteredProjects
	v.updateTable()
	v.projects = originalProjects
}

// GetProjectStatus returns detailed status of a project
func (v *ProjectListView) GetProjectStatus(project models.ComposeProject) (string, error) {
	output, err := docker.ExecuteCaptured("compose", "-p", project.Name, "ps", "--format", "json")
	if err != nil {
		return "", fmt.Errorf("failed to get project status: %w", err)
	}
	return string(output), nil
}

// GetProjectLogs retrieves logs for a project
func (v *ProjectListView) GetProjectLogs(project models.ComposeProject, tail int) (string, error) {
	args := []string{"compose", "-p", project.Name, "logs"}
	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}

	output, err := docker.ExecuteCaptured(args...)
	if err != nil {
		return "", fmt.Errorf("failed to get project logs: %w", err)
	}
	return string(output), nil
}
