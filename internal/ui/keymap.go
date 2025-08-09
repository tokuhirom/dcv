package ui

import (
	"log/slog"
)

func (m *Model) initializeKeyHandlers() {
	m.globalHandlers = []KeyConfig{
		{[]string{"q"}, "quit", m.CmdQuit},
		{[]string{":"}, "command mode", m.CmdCommandMode},

		{[]string{"1"}, "docker ps", m.CmdPS},
		{[]string{"2"}, "project list", m.CmdComposeLS},
		{[]string{"3"}, "docker images", m.CmdImages},
		{[]string{"4"}, "docker networks", m.CmdNetworkLs},
		{[]string{"5"}, "docker volumes", m.CmdVolumeLs},
		{[]string{"6"}, "stats", m.CmdStats},
	}
	m.globalKeymap = m.createKeymap(m.globalHandlers)

	containerOperations := []KeyConfig{
		{[]string{"f"}, "browse files", m.CmdFileBrowse},
		{[]string{"!"}, "exec /bin/sh", m.CmdShell},
		{[]string{"i"}, "inspect", m.CmdInspect},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"a"}, "toggle all", m.CmdToggleAll},
		{[]string{"K"}, "kill", m.CmdKill},
		{[]string{"S"}, "stop", m.CmdStop},
		{[]string{"U"}, "start", m.CmdStart},
		{[]string{"R"}, "restart", m.CmdRestart},
		{[]string{"P"}, "pause/unpause", m.CmdPause},
		{[]string{"D"}, "delete", m.CmdDelete},
	}

	// Docker Container List View
	// `docker ps`
	m.dockerContainerListViewHandlers = append([]KeyConfig{
		{[]string{"up", "k"}, "move up", m.CmdUp},
		{[]string{"down", "j"}, "move down", m.CmdDown},
		{[]string{"enter"}, "view logs", m.CmdLog},
		{[]string{"esc"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},

		{[]string{"d"}, "entering DinD", m.CmdDind},
	}, containerOperations...)
	m.dockerListViewKeymap = m.createKeymap(m.dockerContainerListViewHandlers)

	// Compose Process List View
	m.composeProcessListViewHandlers = append([]KeyConfig{
		{[]string{"up", "k"}, "move up", m.CmdUp},
		{[]string{"down", "j"}, "move down", m.CmdDown},
		{[]string{"enter"}, "view logs", m.CmdLog},
		{[]string{"esc"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},

		// TODO: move to containerOperations
		{[]string{"t"}, "top", m.CmdTop},
		// TODO: support compose stats

		{[]string{"d"}, "entering DinD", m.CmdDind},
	}, containerOperations...)
	m.composeProcessListViewKeymap = m.createKeymap(m.composeProcessListViewHandlers)

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
		{[]string{"esc"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
		{[]string{"ctrl+c"}, "cancel", m.CmdCancel},
	}
	m.logViewKeymap = m.createKeymap(m.logViewHandlers)

	// Dind Process List View
	m.dindListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.CmdUp},
		{[]string{"down", "j"}, "move down", m.CmdDown},
		{[]string{"enter"}, "view logs", m.CmdLog},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"esc"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},

		{[]string{"i"}, "inspect", m.CmdInspect},
		// TODO: support file browser
	}
	m.dindListViewKeymap = m.createKeymap(m.dindListViewHandlers)

	// Top View
	m.topViewHandlers = []KeyConfig{
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"esc"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.topViewKeymap = m.createKeymap(m.topViewHandlers)

	// Stats View
	m.statsViewHandlers = []KeyConfig{
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"esc"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.statsViewKeymap = m.createKeymap(m.statsViewHandlers)

	// Project List View
	m.composeProjectListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.CmdUp},
		{[]string{"down", "j"}, "move down", m.CmdDown},
		{[]string{"enter"}, "select project", m.CmdSelectProject},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.composeProjectListViewKeymap = m.createKeymap(m.composeProjectListViewHandlers)

	// Image List View
	m.imageListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.CmdUp},
		{[]string{"down", "j"}, "move down", m.CmdDown},
		{[]string{"i"}, "inspect", m.CmdInspect},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"a"}, "toggle all", m.CmdToggleAll},
		{[]string{"D"}, "delete", m.CmdDelete},
		{[]string{"esc"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.imageListViewKeymap = m.createKeymap(m.imageListViewHandlers)

	// Network List View
	m.networkListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.CmdUp},
		{[]string{"down", "j"}, "move down", m.CmdDown},
		{[]string{"i"}, "inspect", m.CmdInspect},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"D"}, "delete", m.CmdDelete},
		{[]string{"esc"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.networkListViewKeymap = m.createKeymap(m.networkListViewHandlers)

	// Volume List View
	m.volumeListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.CmdUp},
		{[]string{"down", "j"}, "move down", m.CmdDown},
		{[]string{"i"}, "inspect", m.CmdInspect},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"D"}, "delete", m.CmdDelete},
		{[]string{"esc"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.volumeListViewKeymap = m.createKeymap(m.volumeListViewHandlers)

	// File Browser View
	m.fileBrowserHandlers = []KeyConfig{
		{[]string{"up", "k"}, "move up", m.CmdUp},
		{[]string{"down", "j"}, "move down", m.CmdDown},
		{[]string{"enter"}, "open", m.CmdOpenFileOrDirectory},
		{[]string{"u"}, "parent directory", m.CmdGoToParentDirectory},
		{[]string{"r"}, "refresh", m.CmdRefresh},
		{[]string{"esc"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.fileBrowserKeymap = m.createKeymap(m.fileBrowserHandlers)

	// File Content View
	m.fileContentHandlers = []KeyConfig{
		{[]string{"up", "k"}, "scroll up", m.CmdUp},
		{[]string{"down", "j"}, "scroll down", m.CmdDown},
		{[]string{"G"}, "go to end", m.CmdGoToEnd},
		{[]string{"g"}, "go to start", m.CmdGoToStart},
		{[]string{"esc"}, "back", m.CmdBack},
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
		{[]string{"esc"}, "back", m.CmdBack},
		{[]string{"?"}, "help", m.CmdHelp},
	}
	m.inspectViewKeymap = m.createKeymap(m.inspectViewHandlers)

	// Help View
	m.helpViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "scroll up", m.CmdUp},
		{[]string{"down", "j"}, "scroll down", m.CmdDown},
		{[]string{"esc"}, "back", m.CmdBack},
	}
	m.helpViewKeymap = m.createKeymap(m.helpViewHandlers)

	// Command Execution View
	m.commandExecHandlers = []KeyConfig{
		{[]string{"up", "k"}, "scroll up", m.CmdUp},
		{[]string{"down", "j"}, "scroll down", m.CmdDown},
		{[]string{"G"}, "go to end", m.CmdGoToEnd},
		{[]string{"g"}, "go to start", m.CmdGoToStart},
		{[]string{"ctrl+c"}, "cancel", m.CmdCancel},
		{[]string{"esc"}, "back", m.CmdBack},
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
