package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

func (m *Model) SelectUpContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer > 0 {
		m.selectedContainer--
	}
	return m, nil
}

func (m *Model) SelectDownContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers)-1 {
		m.selectedContainer++
	}
	return m, nil
}

func (m *Model) ShowComposeLog(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	// TODO: abstract this to a common function, dind log and container log and compose log...
	if m.selectedContainer < len(m.containers) {
		process := m.containers[m.selectedContainer]
		m.containerName = process.Name
		m.isDindLog = false
		m.currentView = LogView
		m.logs = []string{}
		m.logScrollY = 0
		// Use service name for docker compose logs
		return m, streamLogs(m.dockerClient, process.ID, false, "")
	}
	return m, nil
}

func (m *Model) ShowDindProcessList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		container := m.containers[m.selectedContainer]
		if container.IsDind() {
			m.currentDindHost = container.Name
			m.currentDindContainerID = container.ID
			m.currentView = DindProcessListView
			m.loading = true
			return m, loadDindContainers(m.dockerClient, container.ID)
		}
	}
	return m, nil
}

func (m *Model) RefreshProcessList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.loading = true
	return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
}

func (m *Model) ToggleAllContainers(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.showAll = !m.showAll
	m.loading = true
	return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
}

func (m *Model) ShowStatsView(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.currentView = StatsView
	m.loading = true
	return m, loadStats(m.dockerClient)
}

func (m *Model) ShowTopView(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		container := m.containers[m.selectedContainer]
		m.topService = container.Service
		m.currentView = TopView
		m.loading = true
		return m, loadTop(m.dockerClient, m.projectName, container.Service)
	}
	return m, nil
}

func (m *Model) KillContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		container := m.containers[m.selectedContainer]
		m.loading = true
		return m, killService(m.dockerClient, container.ID)
	}
	return m, nil
}

func (m *Model) StopContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		container := m.containers[m.selectedContainer]
		m.loading = true
		return m, stopService(m.dockerClient, container.ID)
	}
	return m, nil
}

func (m *Model) UpService(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		container := m.containers[m.selectedContainer]
		m.loading = true
		return m, startService(m.dockerClient, container.Service)
	}
	return m, nil
}

func (m *Model) RestartContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		container := m.containers[m.selectedContainer]
		m.loading = true
		return m, restartService(m.dockerClient, container.ID)
	}
	return m, nil
}

func (m *Model) DeleteContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		container := m.containers[m.selectedContainer]
		// Only allow removing stopped containers
		if !strings.Contains(container.Status, "Up") && !strings.Contains(container.State, "running") {
			m.loading = true
			return m, removeService(m.dockerClient, container.ID)
		}
	}
	return m, nil
}

func (m *Model) DeployProject(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		m.loading = true
		return m, up(m.dockerClient, m.projectName)
	}
	return m, nil
}

func (m *Model) ShowProjectList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.currentView = ProjectListView
	m.loading = true
	return m, loadProjects(m.dockerClient)
}
