package ui

import (
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/models"
)

type ComposeProcessListViewModel struct {
	// Process list state
	composeContainers []models.ComposeContainer
	selectedContainer int
	showAll           bool // Toggle to show all composeContainers including stopped ones
}

func (m *ComposeProcessListViewModel) Load(model *Model, project models.ComposeProject) tea.Cmd {
	model.projectName = project.Name
	model.currentView = ComposeProcessListView
	model.loading = true
	return loadProcesses(model.dockerClient, model.projectName, m.showAll)
}

func (m *ComposeProcessListViewModel) render(model *Model, availableHeight int) string {
	slog.Info("Rendering container list",
		slog.Int("selectedContainer", m.selectedContainer),
		slog.Int("numContainers", len(m.composeContainers)))

	// Empty state
	if len(m.composeContainers) == 0 {
		var s strings.Builder
		s.WriteString("\nNo containers found.\n")
		s.WriteString("\nPress u to start services or p to switch to project list\n")
		return s.String()
	}

	// Create table with fixed widths
	columns := []table.Column{
		{Title: "SERVICE", Width: 20},
		{Title: "IMAGE", Width: 30},
		{Title: "STATUS", Width: 20},
		{Title: "PORTS", Width: model.width - 75},
	}

	rows := make([]table.Row, 0, len(m.composeContainers))
	// Add rows with width control
	for _, container := range m.composeContainers {
		// Service name with dind indicator
		service := container.Service
		if container.IsDind() {
			service = dindStyle.Render("⬢ ") + service
		}

		// Truncate image name if too long
		image := container.Image
		if len(image) > 30 {
			image = image[:27] + "..."
		}

		// Status with color
		status := container.GetStatus()
		if strings.Contains(status, "Up") || strings.Contains(status, "running") {
			status = statusUpStyle.Render(status)
		} else {
			status = statusDownStyle.Render(status)
		}

		// Truncate ports if too long
		ports := container.GetPortsString()
		if len(ports) > 40 {
			ports = ports[:37] + "..."
		}

		rows = append(rows, table.Row{service, image, status, ports})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(availableHeight-2),
		table.WithFocused(true),
	)

	// Apply styles
	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	styles.Selected = selectedStyle
	styles.Cell = styles.Cell.
		BorderForeground(lipgloss.Color("240"))
	t.SetStyles(styles)

	// Set cursor position
	if m.selectedContainer < len(rows) {
		t.MoveDown(m.selectedContainer)
	}

	return t.View()
}

func (m *ComposeProcessListViewModel) HandleUp() tea.Cmd {
	if m.selectedContainer > 0 {
		m.selectedContainer--
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleDown() tea.Cmd {
	if m.selectedContainer < len(m.composeContainers)-1 {
		m.selectedContainer++
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleLog(model *Model) tea.Cmd {
	if m.selectedContainer < len(m.composeContainers) {
		composeContainer := m.composeContainers[m.selectedContainer]
		model.logViewModel.SwitchToLogView(model, composeContainer.Name)
		return model.logViewModel.StreamComposeLogs(model, composeContainer)
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleToggleAll(model *Model) tea.Cmd {
	m.showAll = !m.showAll
	model.loading = true
	return loadProcesses(model.dockerClient, model.projectName, m.showAll)
}

func (m *ComposeProcessListViewModel) HandleTop(model *Model) tea.Cmd {
	if m.selectedContainer < len(m.composeContainers) {
		container := m.composeContainers[m.selectedContainer]
		return model.topViewModel.Load(model, model.projectName, container.Service)
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleDindProcessList(model *Model) tea.Cmd {
	if m.selectedContainer < len(m.composeContainers) {
		container := m.composeContainers[m.selectedContainer]
		if container.IsDind() {
			return model.dindProcessListViewModel.Load(model, container)
		}
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleKill(model *Model) tea.Cmd {
	if m.selectedContainer < len(m.composeContainers) {
		container := m.composeContainers[m.selectedContainer]
		return model.commandExecutionViewModel.ExecuteCommand(model, "kill", container.ID)
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleStop(model *Model) tea.Cmd {
	if m.selectedContainer < len(m.composeContainers) {
		container := m.composeContainers[m.selectedContainer]
		return model.commandExecutionViewModel.ExecuteCommand(model, "stop", container.ID)
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleStart(model *Model) tea.Cmd {
	if m.selectedContainer < len(m.composeContainers) {
		container := m.composeContainers[m.selectedContainer]
		return model.commandExecutionViewModel.ExecuteCommand(model, "start", container.ID)
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleRestart(model *Model) tea.Cmd {
	if m.selectedContainer < len(m.composeContainers) {
		container := m.composeContainers[m.selectedContainer]
		return model.commandExecutionViewModel.ExecuteCommand(model, "restart", container.ID)
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleRemove(model *Model) tea.Cmd {
	if m.selectedContainer < len(m.composeContainers) {
		container := m.composeContainers[m.selectedContainer]
		// Only allow removing stopped composeContainers
		if !strings.Contains(container.GetStatus(), "Up") && !strings.Contains(container.State, "running") {
			model.loading = true
			return removeService(model.dockerClient, container.ID)
		}
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleFileBrowse(model *Model) tea.Cmd {
	if m.selectedContainer < len(m.composeContainers) {
		container := m.composeContainers[m.selectedContainer]
		return model.fileBrowserViewModel.Load(model, container.ID, container.Name)
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleShell(model *Model) tea.Cmd {
	if m.selectedContainer < len(m.composeContainers) {
		container := m.composeContainers[m.selectedContainer]
		// Default to /bin/sh as it's most commonly available
		return executeInteractiveCommand(container.ID, []string{"/bin/sh"})
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleInspect(model *Model) tea.Cmd {
	if m.selectedContainer < len(m.composeContainers) {
		container := m.composeContainers[m.selectedContainer]
		return model.inspectViewModel.InspectContainer(model, container.ID)
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}
