package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
)

// fileContentLoadedMsg contains the loaded file content
type fileContentLoadedMsg struct {
	content string
	path    string
	err     error
}

type FileContentViewModel struct {
	container   *docker.Container
	content     string
	contentPath string
	scrollY     int
}

// Update handles messages for the file content view
func (m *FileContentViewModel) Update(model *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case fileContentLoadedMsg:
		model.loading = false
		if msg.err != nil {
			model.err = msg.err
			return model, nil
		} else {
			model.err = nil
		}

		m.Loaded(msg.content, msg.path)
		return model, nil
	default:
		return model, nil
	}
}

// render renders the file content view
func (m *FileContentViewModel) render(model *Model) string {
	if model.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		return errorStyle.Render(fmt.Sprintf("Error: %v", model.err))
	}

	v := viewport.New(model.width, model.Height-4)
	v.SetContent(m.content)
	v.ScrollDown(m.scrollY)
	return v.View()
}

func (m *FileContentViewModel) LoadContainer(model *Model, container *docker.Container, path string) tea.Cmd {
	model.SwitchView(FileContentView)
	model.loading = true
	m.scrollY = 0
	m.container = container

	return func() tea.Msg {
		// Try docker cp first (works for files without needing exec permissions)
		var args []string
		if container.IsDind() {
			// For DinD: docker exec <host> docker cp <container>:<path> -
			args = []string{"exec", container.HostContainerID(), "docker", "cp", fmt.Sprintf("%s:%s", container.ContainerID(), path), "-"}
		} else {
			// For normal containers: docker cp <container>:<path> -
			args = []string{"cp", fmt.Sprintf("%s:%s", container.ContainerID(), path), "-"}
		}

		output, err := docker.ExecuteCaptured(args...)
		if err == nil {
			return fileContentLoadedMsg{
				content: string(output),
				path:    path,
				err:     nil,
			}
		}

		// Fallback to cat if docker cp fails (e.g., for special files like /proc/*)
		args = container.OperationArgs("exec", "cat", path)
		output, err = docker.ExecuteCaptured(args...)
		if err != nil {
			return fileContentLoadedMsg{
				content: "",
				path:    path,
				err:     fmt.Errorf("failed to read file: %w", err),
			}
		}

		return fileContentLoadedMsg{
			content: string(output),
			path:    path,
			err:     nil,
		}
	}
}

func (m *FileContentViewModel) HandleUp() tea.Cmd {
	if m.scrollY > 0 {
		m.scrollY--
	}
	return nil
}

func (m *FileContentViewModel) HandleDown(height int) tea.Cmd {
	lines := strings.Split(m.content, "\n")
	maxScroll := len(lines) - (height - 5)
	if m.scrollY < maxScroll && maxScroll > 0 {
		m.scrollY++
	}
	return nil
}

func (m *FileContentViewModel) HandleGoToBeginning() tea.Cmd {
	m.scrollY = 0
	return nil
}

func (m *FileContentViewModel) HandleGoToEnd(height int) tea.Cmd {
	lines := strings.Split(m.content, "\n")
	maxScroll := len(lines) - (height - 5)
	if maxScroll > 0 {
		m.scrollY = maxScroll
	}
	return nil
}

func (m *FileContentViewModel) HandlePageUp(height int) tea.Cmd {
	pageSize := height - 5
	if m.scrollY > pageSize {
		m.scrollY -= pageSize
	} else {
		m.scrollY = 0
	}
	return nil
}

func (m *FileContentViewModel) HandlePageDown(height int) tea.Cmd {
	lines := strings.Split(m.content, "\n")
	maxScroll := len(lines) - (height - 5)
	pageSize := height - 5

	if m.scrollY+pageSize < maxScroll {
		m.scrollY += pageSize
	} else if maxScroll > 0 {
		m.scrollY = maxScroll
	}
	return nil
}

func (m *FileContentViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	m.content = ""
	m.contentPath = ""
	m.contentPath = ""
	m.scrollY = 0
	return nil
}

func (m *FileContentViewModel) Title() string {
	containerTitle := ""
	if m.container != nil {
		containerTitle = m.container.Title()
	}
	return fmt.Sprintf("File: [%d/%d] %s [%s] ",
		m.scrollY, len(strings.Split(m.content, "\n")),
		m.contentPath,
		containerTitle,
	)
}

func (m *FileContentViewModel) Loaded(content string, path string) {
	m.content = content
	m.contentPath = path
	m.scrollY = 0
}
