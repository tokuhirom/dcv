package ui

import (
	"log/slog"
)

func (m *Model) initializeKeyHandlers() {
	m.globalHandlers = []KeyConfig{
		{[]string{"q"}, "quit", m.CmdQuit},
		{[]string{":"}, "command mode", m.CmdCommandMode},

		{[]string{"1"}, "docker ps", m.ShowDockerContainerList},
		{[]string{"2"}, "project list", m.ShowProjectList},
		{[]string{"3"}, "docker images", m.ShowImageList},
		{[]string{"4"}, "docker networks", m.ShowNetworkList},
		{[]string{"5"}, "docker volumes", m.ShowVolumeList},
	}
	m.globalKeymap = m.createKeymap(m.globalHandlers)

	// Process List View
	m.processListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.CmdUp},
		{[]string{"down", "j"}, "move down", m.CmdDown},
		{[]string{"enter"}, "view logs", m.CmdLog},
		{[]string{"d"}, "dind composeContainers", m.ShowDindProcessList},
		{[]string{"f"}, "browse files", m.CmdFileBrowse},
		{[]string{"!"}, "exec /bin/sh", m.CmdShell},
		{[]string{"i"}, "inspect", m.CmdInspect},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"a"}, "toggle all", m.CmdToggleAll},
		{[]string{"s"}, "stats", m.ShowStatsView},
		{[]string{"t"}, "top", m.CmdTop},
		{[]string{"K"}, "kill", m.CmdKill},
		{[]string{"S"}, "stop", m.CmdStop},
		{[]string{"U"}, "start", m.CmdStart},
		{[]string{"R"}, "restart", m.CmdRestart},
		{[]string{"P"}, "pause/unpause", m.CmdPause},
		{[]string{"D"}, "remove", m.CmdRemove},
		{[]string{"u"}, "up -d", m.DeployProject},
		{[]string{"x"}, "down", m.DownProject},
		{[]string{"?"}, "help", m.CmdHelp},
		{[]string{"esc", "q"}, "back", m.CmdBack},
	}
	m.processListViewKeymap = m.createKeymap(m.processListViewHandlers)

	// Log View
	m.logViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "scroll up", m.CmdUp},
		{[]string{"down", "j"}, "scroll down", m.CmdDown},
		{[]string{"G"}, "go to end", m.CmdGoToEnd},
		{[]string{"g"}, "go to start", m.CmdGoToStart},
		{[]string{"/"}, "search", m.CmdSearch},
		{[]string{"n"}, "next match", m.CmdNextSearchResult},
		{[]string{"N"}, "prev match", m.CmdPrevSearchResult},
		{[]string{"f"}, "filter", m.CmdFilter},
		{[]string{"esc", "q"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
		{[]string{"ctrl+c"}, "cancel", m.CmdCancel},
	}
	m.logViewKeymap = m.createKeymap(m.logViewHandlers)

	// Dind Process List View
	m.dindListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.SelectUpDindContainer},
		{[]string{"down", "j"}, "move down", m.SelectDownDindContainer},
		{[]string{"enter"}, "view logs", m.ShowDindLog},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"esc"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.dindListViewKeymap = m.createKeymap(m.dindListViewHandlers)

	// Top View
	m.topViewHandlers = []KeyConfig{
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"esc", "q"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.topViewKeymap = m.createKeymap(m.topViewHandlers)

	// Stats View
	m.statsViewHandlers = []KeyConfig{
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"esc", "q"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.statsViewKeymap = m.createKeymap(m.statsViewHandlers)

	// Project List View
	m.composeProjectListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.CmdUp},
		{[]string{"down", "j"}, "move down", m.CmdDown},
		{[]string{"enter"}, "select project", m.SelectProject},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.composeProjectListViewKeymap = m.createKeymap(m.composeProjectListViewHandlers)

	// Docker Container List View
	m.dockerContainerListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.CmdUp},
		{[]string{"down", "j"}, "move down", m.CmdDown},
		{[]string{"enter"}, "view logs", m.CmdLog},
		{[]string{"f"}, "browse files", m.CmdFileBrowse},
		{[]string{"!"}, "exec /bin/sh", m.CmdShell},
		{[]string{"i"}, "inspect", m.ShowDockerInspect},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"a"}, "toggle all", m.ToggleAllDockerContainers},
		{[]string{"K"}, "kill", m.CmdKill},
		{[]string{"S"}, "stop", m.CmdStop},
		{[]string{"U"}, "start", m.CmdStart},
		{[]string{"R"}, "restart", m.CmdRestart},
		{[]string{"P"}, "pause/unpause", m.CmdPause},
		{[]string{"D"}, "remove", m.CmdRemove},
		{[]string{"esc", "q"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.dockerListViewKeymap = m.createKeymap(m.dockerContainerListViewHandlers)

	// Image List View
	m.imageListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.SelectUpImage},
		{[]string{"down", "j"}, "move down", m.SelectDownImage},
		{[]string{"i"}, "inspect", m.ShowImageInspect},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"a"}, "toggle all", m.ToggleAllImages},
		{[]string{"D"}, "remove", m.DeleteImage},
		{[]string{"F"}, "force remove", m.ForceDeleteImage},
		{[]string{"esc", "q"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.imageListViewKeymap = m.createKeymap(m.imageListViewHandlers)

	// Network List View
	m.networkListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.SelectUpNetwork},
		{[]string{"down", "j"}, "move down", m.SelectDownNetwork},
		{[]string{"i"}, "inspect", m.ShowNetworkInspect},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"D"}, "remove", m.DeleteNetwork},
		{[]string{"esc", "q"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.networkListViewKeymap = m.createKeymap(m.networkListViewHandlers)

	// Volume List View
	m.volumeListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.SelectUpVolume},
		{[]string{"down", "j"}, "move down", m.SelectDownVolume},
		{[]string{"i"}, "inspect", m.ShowVolumeInspect},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"D"}, "remove", m.DeleteVolume},
		{[]string{"F"}, "force remove", m.ForceDeleteVolume},
		{[]string{"esc", "q"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.volumeListViewKeymap = m.createKeymap(m.volumeListViewHandlers)

	// File Browser View
	m.fileBrowserHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.CmdUp},
		{[]string{"down", "j"}, "move down", m.CmdDown},
		{[]string{"enter"}, "open", m.OpenFileOrDirectory},
		{[]string{"u"}, "parent directory", m.GoToParentDirectory},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"esc", "q"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.fileBrowserKeymap = m.createKeymap(m.fileBrowserHandlers)

	// File Content View
	m.fileContentHandlers = []KeyConfig{
		{[]string{"up", "k"}, "scroll up", m.ScrollFileUp},
		{[]string{"down", "j"}, "scroll down", m.ScrollFileDown},
		{[]string{"G"}, "go to end", m.GoToFileEnd},
		{[]string{"g"}, "go to start", m.GoToFileStart},
		{[]string{"esc", "q"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.fileContentKeymap = m.createKeymap(m.fileContentHandlers)

	// Inspect View
	m.inspectViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "scroll up", m.CmdUp},
		{[]string{"down", "j"}, "scroll down", m.CmdDown},
		{[]string{"G"}, "go to end", m.CmdGoToEnd},
		{[]string{"g"}, "go to start", m.CmdGoToStart},
		{[]string{"/"}, "search", m.CmdSearch},
		{[]string{"n"}, "next match", m.CmdNextSearchResult},
		{[]string{"N"}, "prev match", m.CmdPrevSearchResult},
		{[]string{"esc", "q"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.inspectViewKeymap = m.createKeymap(m.inspectViewHandlers)

	// Help View
	m.helpViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "scroll up", m.ScrollHelpUp},
		{[]string{"down", "j"}, "scroll down", m.ScrollHelpDown},
		{[]string{"esc", "q"}, "back", m.CmdBack},
	}
	m.helpViewKeymap = m.createKeymap(m.helpViewHandlers)

	// Command Execution View
	m.commandExecHandlers = []KeyConfig{
		{[]string{"up", "k"}, "scroll up", m.CmdUp},
		{[]string{"down", "j"}, "scroll down", m.CmdDown},
		{[]string{"G"}, "go to end", m.CmdGoToEnd},
		{[]string{"g"}, "go to start", m.CmdGoToStart},
		{[]string{"ctrl+c"}, "cancel", m.CmdCancel},
		{[]string{"esc", "q"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.commandExecKeymap = m.createKeymap(m.commandExecHandlers)

	// Initialize command registry
	m.initCommandRegistry()

	slog.Info("Initialized all view keymaps and command registry")
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
