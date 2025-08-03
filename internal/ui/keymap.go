package ui

import (
	"log/slog"
)

func (m *Model) initializeKeyHandlers() {
	// Process List View
	m.processListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.SelectUpContainer},
		{[]string{"down", "j"}, "move down", m.SelectDownContainer},
		{[]string{"enter"}, "view logs", m.ShowComposeLog},
		{[]string{"d"}, "dind composeContainers", m.ShowDindProcessList},
		{[]string{"1"}, "docker ps", m.ShowDockerContainerList},
		{[]string{"3"}, "docker images", m.ShowImageList},
		{[]string{"4"}, "docker networks", m.ShowNetworkList},
		{[]string{"f"}, "browse files", m.ShowFileBrowser},
		{[]string{"!"}, "exec /bin/sh", m.ExecuteShell},
		{[]string{"I"}, "inspect", m.ShowInspect},
		{[]string{"r"}, "refresh", m.RefreshProcessList},
		{[]string{"a"}, "toggle all", m.ToggleAllContainers},
		{[]string{"s"}, "stats", m.ShowStatsView},
		{[]string{"t"}, "top", m.ShowTopView},
		{[]string{"K"}, "kill", m.KillContainer},
		{[]string{"S"}, "stop", m.StopContainer},
		{[]string{"U"}, "start", m.UpService},
		{[]string{"R"}, "restart", m.RestartContainer},
		{[]string{"P"}, "pause/unpause", m.PauseContainer},
		{[]string{"D"}, "remove", m.DeleteContainer},
		{[]string{"u"}, "up -d", m.DeployProject},
		{[]string{"x"}, "down", m.DownProject},
		{[]string{"2"}, "project list", m.ShowProjectList},
		{[]string{"?"}, "help", m.ShowHelp},
	}
	m.processListViewKeymap = m.createKeymap(m.processListViewHandlers)

	// Log View
	m.logViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "scroll up", m.ScrollLogUp},
		{[]string{"down", "j"}, "scroll down", m.ScrollLogDown},
		{[]string{"G"}, "go to end", m.GoToLogEnd},
		{[]string{"g"}, "go to start", m.GoToLogStart},
		{[]string{"/"}, "search", m.StartSearch},
		{[]string{"n"}, "next match", m.NextSearchResult},
		{[]string{"N"}, "prev match", m.PrevSearchResult},
		{[]string{"esc"}, "back", m.BackFromLogView},
		{[]string{"?"}, "help", m.ShowHelp},
	}
	m.logViewKeymap = m.createKeymap(m.logViewHandlers)

	// Dind Process List View
	m.dindListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.SelectUpDindContainer},
		{[]string{"down", "j"}, "move down", m.SelectDownDindContainer},
		{[]string{"enter"}, "view logs", m.ShowDindLog},
		{[]string{"r"}, "refresh", m.RefreshDindList},
		{[]string{"1"}, "docker ps", m.ShowDockerContainerList},
		{[]string{"2"}, "project list", m.ShowProjectList},
		{[]string{"3"}, "docker images", m.ShowImageList},
		{[]string{"4"}, "docker networks", m.ShowNetworkList},
		{[]string{"esc"}, "back", m.BackToDindList},
		{[]string{"?"}, "help", m.ShowHelp},
	}
	m.dindListViewKeymap = m.createKeymap(m.dindListViewHandlers)

	// Top View
	m.topViewHandlers = []KeyConfig{
		{[]string{"r"}, "refresh", m.RefreshTop},
		{[]string{"1"}, "docker ps", m.ShowDockerContainerList},
		{[]string{"2"}, "project list", m.ShowProjectList},
		{[]string{"3"}, "docker images", m.ShowImageList},
		{[]string{"4"}, "docker networks", m.ShowNetworkList},
		{[]string{"esc", "q"}, "back", m.BackToProcessList},
		{[]string{"?"}, "help", m.ShowHelp},
	}
	m.topViewKeymap = m.createKeymap(m.topViewHandlers)

	// Stats View
	m.statsViewHandlers = []KeyConfig{
		{[]string{"r"}, "refresh", m.RefreshStats},
		{[]string{"1"}, "docker ps", m.ShowDockerContainerList},
		{[]string{"2"}, "project list", m.ShowProjectList},
		{[]string{"3"}, "docker images", m.ShowImageList},
		{[]string{"4"}, "docker networks", m.ShowNetworkList},
		{[]string{"esc", "q"}, "back", m.BackToProcessList},
		{[]string{"?"}, "help", m.ShowHelp},
	}
	m.statsViewKeymap = m.createKeymap(m.statsViewHandlers)

	// Project List View
	m.projectListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.SelectUpProject},
		{[]string{"down", "j"}, "move down", m.SelectDownProject},
		{[]string{"enter"}, "select project", m.SelectProject},
		{[]string{"r"}, "refresh", m.RefreshProjects},
		{[]string{"1"}, "docker ps", m.ShowDockerContainerList},
		{[]string{"3"}, "docker images", m.ShowImageList},
		{[]string{"4"}, "docker networks", m.ShowNetworkList},
		{[]string{"?"}, "help", m.ShowHelp},
	}
	m.projectListViewKeymap = m.createKeymap(m.projectListViewHandlers)

	// Docker Container List View
	m.dockerListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.SelectUpDockerContainer},
		{[]string{"down", "j"}, "move down", m.SelectDownDockerContainer},
		{[]string{"enter"}, "view logs", m.ShowDockerLog},
		{[]string{"f"}, "browse files", m.ShowDockerFileBrowser},
		{[]string{"!"}, "exec /bin/sh", m.ExecuteDockerShell},
		{[]string{"I"}, "inspect", m.ShowDockerInspect},
		{[]string{"r"}, "refresh", m.RefreshDockerList},
		{[]string{"a"}, "toggle all", m.ToggleAllDockerContainers},
		{[]string{"K"}, "kill", m.KillDockerContainer},
		{[]string{"S"}, "stop", m.StopDockerContainer},
		{[]string{"U"}, "start", m.StartDockerContainer},
		{[]string{"R"}, "restart", m.RestartDockerContainer},
		{[]string{"P"}, "pause/unpause", m.PauseDockerContainer},
		{[]string{"D"}, "remove", m.DeleteDockerContainer},
		{[]string{"2"}, "project list", m.ShowProjectList},
		{[]string{"3"}, "docker images", m.ShowImageList},
		{[]string{"4"}, "docker networks", m.ShowNetworkList},
		{[]string{"esc", "q"}, "back", m.BackFromDockerList},
		{[]string{"?"}, "help", m.ShowHelp},
	}
	m.dockerListViewKeymap = m.createKeymap(m.dockerListViewHandlers)

	// Image List View
	m.imageListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.SelectUpImage},
		{[]string{"down", "j"}, "move down", m.SelectDownImage},
		{[]string{"I"}, "inspect", m.ShowImageInspect},
		{[]string{"r"}, "refresh", m.RefreshImageList},
		{[]string{"a"}, "toggle all", m.ToggleAllImages},
		{[]string{"D"}, "remove", m.DeleteImage},
		{[]string{"F"}, "force remove", m.ForceDeleteImage},
		{[]string{"1"}, "docker ps", m.ShowDockerContainerList},
		{[]string{"2"}, "project list", m.ShowProjectList},
		{[]string{"4"}, "docker networks", m.ShowNetworkList},
		{[]string{"esc", "q"}, "back", m.BackFromImageList},
		{[]string{"?"}, "help", m.ShowHelp},
	}
	m.imageListViewKeymap = m.createKeymap(m.imageListViewHandlers)

	// Network List View
	m.networkListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.SelectUpNetwork},
		{[]string{"down", "j"}, "move down", m.SelectDownNetwork},
		{[]string{"enter"}, "inspect", m.ShowNetworkInspect},
		{[]string{"r"}, "refresh", m.RefreshNetworkList},
		{[]string{"D"}, "remove", m.DeleteNetwork},
		{[]string{"1"}, "docker ps", m.ShowDockerContainerList},
		{[]string{"2"}, "project list", m.ShowProjectList},
		{[]string{"3"}, "docker images", m.ShowImageList},
		{[]string{"esc", "q"}, "back", m.BackFromNetworkList},
		{[]string{"?"}, "help", m.ShowHelp},
	}
	m.networkListViewKeymap = m.createKeymap(m.networkListViewHandlers)

	// File Browser View
	m.fileBrowserHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.SelectUpFile},
		{[]string{"down", "j"}, "move down", m.SelectDownFile},
		{[]string{"enter"}, "open", m.OpenFileOrDirectory},
		{[]string{"u"}, "parent directory", m.GoToParentDirectory},
		{[]string{"r"}, "refresh", m.RefreshFiles},
		{[]string{"esc", "q"}, "back", m.BackFromFileBrowser},
		{[]string{"?"}, "help", m.ShowHelp},
	}
	m.fileBrowserKeymap = m.createKeymap(m.fileBrowserHandlers)

	// File Content View
	m.fileContentHandlers = []KeyConfig{
		{[]string{"up", "k"}, "scroll up", m.ScrollFileUp},
		{[]string{"down", "j"}, "scroll down", m.ScrollFileDown},
		{[]string{"G"}, "go to end", m.GoToFileEnd},
		{[]string{"g"}, "go to start", m.GoToFileStart},
		{[]string{"esc", "q"}, "back", m.BackFromFileContent},
		{[]string{"?"}, "help", m.ShowHelp},
	}
	m.fileContentKeymap = m.createKeymap(m.fileContentHandlers)

	// Inspect View
	m.inspectViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "scroll up", m.ScrollInspectUp},
		{[]string{"down", "j"}, "scroll down", m.ScrollInspectDown},
		{[]string{"G"}, "go to end", m.GoToInspectEnd},
		{[]string{"g"}, "go to start", m.GoToInspectStart},
		{[]string{"esc", "q"}, "back", m.BackFromInspect},
		{[]string{"?"}, "help", m.ShowHelp},
	}
	m.inspectViewKeymap = m.createKeymap(m.inspectViewHandlers)

	// Help View
	m.helpViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "scroll up", m.ScrollHelpUp},
		{[]string{"down", "j"}, "scroll down", m.ScrollHelpDown},
		{[]string{"esc", "q"}, "back", m.BackFromHelp},
	}
	m.helpViewKeymap = m.createKeymap(m.helpViewHandlers)

	slog.Info("Initialized all view keymaps")
}

func (m *Model) createKeymap(configs []KeyConfig) map[string]KeyHandler {
	keymap := make(map[string]KeyHandler)
	for _, config := range configs {
		for _, key := range config.Keys {
			keymap[key] = config.KeyHandler
		}
	}
	return keymap
}
