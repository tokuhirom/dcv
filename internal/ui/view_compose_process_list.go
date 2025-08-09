package ui

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"

	"github.com/tokuhirom/dcv/internal/models"
)

type ComposeProcessListViewModel struct {
	// Process list state
	composeContainers []models.ComposeContainer
	selectedContainer int
	showAll           bool   // Toggle to show all composeContainers including stopped ones
	projectName       string // Current Docker Compose project name
}

func (m *ComposeProcessListViewModel) Load(model *Model, project models.ComposeProject) tea.Cmd {
	m.projectName = project.Name
	model.SwitchView(ComposeProcessListView)
	return m.DoLoad(model)
}

func (m *ComposeProcessListViewModel) DoLoad(model *Model) tea.Cmd {
	model.loading = true
	return func() tea.Msg {
		slog.Info("Loading composeContainers",
			slog.Bool("showAll", m.showAll))
		processes, err := model.dockerClient.Compose(m.projectName).ListContainers(m.showAll)
		return processesLoadedMsg{
			processes: processes,
			err:       err,
		}
	}
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
	return m.DoLoad(model)
}

func (m *ComposeProcessListViewModel) HandleDindProcessList(model *Model) tea.Cmd {
	container := m.GetContainer(model)
	if container == nil {
		slog.Error("Failed to get selected container for DIND process list",
			slog.Any("error", fmt.Errorf("no container selected")))
		return nil
	}

	return model.dindProcessListViewModel.Load(model, container)
}

func (m *ComposeProcessListViewModel) HandleCommandExecution(model *Model, operation string, aggressive bool) tea.Cmd {
	// TODO: deprecate this
	if m.selectedContainer < len(m.composeContainers) {
		container := m.composeContainers[m.selectedContainer]
		return model.commandExecutionViewModel.ExecuteCommand(model, aggressive, operation, container.ID)
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleRemove(model *Model) tea.Cmd {
	if m.selectedContainer < len(m.composeContainers) {
		container := m.composeContainers[m.selectedContainer]
		// Only allow removing stopped containers
		if !strings.Contains(container.GetStatus(), "Up") && !strings.Contains(container.State, "running") {
			// Use CommandExecutionView to show real-time output
			return model.commandExecutionViewModel.ExecuteCommand(model, true, "rm", "-f", container.ID) // rm is aggressive
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

func (m *ComposeProcessListViewModel) HandleShell() tea.Cmd {
	if m.selectedContainer < len(m.composeContainers) {
		container := m.composeContainers[m.selectedContainer]
		// Default to /bin/sh as it's most commonly available
		return func() tea.Msg {
			return executeCommandMsg{
				containerID: container.ID,
				command:     []string{"/bin/sh"},
			}
		}
	}
	return nil
}

func (m *ComposeProcessListViewModel) GetContainer(model *Model) docker.Container {
	if m.selectedContainer < len(m.composeContainers) {
		container := m.composeContainers[m.selectedContainer]
		return docker.NewContainer(model.dockerClient, container.ID, container.Name,
			fmt.Sprintf("%s(project:%s)", container.Service, m.projectName), container.State)
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleInspect(model *Model) tea.Cmd {
	container := m.GetContainer(model)
	if container == nil {
		slog.Error("Failed to get selected container for inspection")
		return nil
	}

	return model.inspectViewModel.Inspect(model,
		fmt.Sprintf("compose process: %s(%s)", container.GetName(), m.projectName),
		func() ([]byte, error) {
			return container.Inspect()
		})
}

func (m *ComposeProcessListViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}

func (m *ComposeProcessListViewModel) Loaded(processes []models.ComposeContainer) {
	m.composeContainers = processes
	if len(m.composeContainers) > 0 && m.selectedContainer >= len(m.composeContainers) {
		m.selectedContainer = 0
	}
}
