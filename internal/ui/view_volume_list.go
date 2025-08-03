package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
)

func (m Model) renderVolumeList() string {
	s := strings.Builder{}
	s.WriteString(titleStyle.Render("📦 Docker Volumes"))
	s.WriteString("\n\n")

	if m.loading {
		s.WriteString("Loading volumes...")
		return s.String()
	}

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())))
		return s.String()
	}

	if len(m.dockerVolumes) == 0 {
		s.WriteString("No volumes found.\n")
		s.WriteString(helpStyle.Render("\nPress 'q' to go back"))
		return s.String()
	}

	// Create table columns
	columns := []table.Column{
		{Title: "Name", Width: 40},
		{Title: "Driver", Width: 10},
		{Title: "Scope", Width: 10},
		{Title: "Size", Width: 12},
		{Title: "Created", Width: 20},
		{Title: "Ref Count", Width: 10},
	}

	// Create table rows
	rows := make([]table.Row, len(m.dockerVolumes))
	for i, volume := range m.dockerVolumes {
		size := formatBytes(volume.Size)
		if volume.Size == 0 {
			size = "-"
		}

		refCount := fmt.Sprintf("%d", volume.RefCount)
		if volume.RefCount == 0 && volume.Size == 0 {
			refCount = "-"
		}

		created := volume.CreatedAt.Format("2006-01-02 15:04:05")
		if volume.CreatedAt.IsZero() {
			created = "-"
		}

		rows[i] = table.Row{
			volume.Name,
			volume.Driver,
			volume.Scope,
			size,
			created,
			refCount,
		}
	}

	// Create table
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(len(rows), m.height-8)),
	)

	tableStyle := table.DefaultStyles()
	tableStyle.Header = tableStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	tableStyle.Selected = tableStyle.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t.SetStyles(tableStyle)
	t.Focus()

	// Move to selected row
	for i := 0; i < m.selectedDockerVolume; i++ {
		t.MoveDown(1)
	}

	s.WriteString(t.View())
	s.WriteString("\n")

	return s.String()
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func loadDockerVolumes(dockerClient *docker.Client) tea.Cmd {
	return func() tea.Msg {
		volumes, err := dockerClient.ListVolumes()
		return dockerVolumesLoadedMsg{volumes: volumes, err: err}
	}
}
