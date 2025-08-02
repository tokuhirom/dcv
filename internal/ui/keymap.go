package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"log/slog"
)

type KeyHandler func(msg tea.KeyMsg) (tea.Model, tea.Cmd)

type KeyConfig struct {
	Keys        []string
	HandlerName string
	KeyHandler  KeyHandler
}

func (m *Model) initializeKeyHandlers() {
	m.processListViewHandlers = []KeyConfig{
		{[]string{"up", "k"}, "up", m.SelectUpContainer},
		{[]string{"down", "j"}, "down", m.SelectDownContainer},
		{[]string{"enter"}, "log", m.ShowComposeLog},
		{[]string{"d"}, "dind", m.ShowDindProcessList},
		{[]string{"r"}, "refresh", m.RefreshProcessList},
		{[]string{"a"}, "toggleAll", m.ToggleAllContainers},
		{[]string{"s"}, "stats", m.ShowStatsView},
		{[]string{"t"}, "top", m.ShowTopView},
		{[]string{"K"}, "kill", m.KillContainer},
		{[]string{"S"}, "stop", m.StopContainer},
		{[]string{"U"}, "up", m.UpService},
		{[]string{"R"}, "restart", m.RestartContainer},
		{[]string{"D"}, "remove", m.DeleteContainer},
	}
	keymap := make(map[string]KeyHandler)
	for _, config := range m.processListViewHandlers {
		for _, key := range config.Keys {
			slog.Info("Registering key handler",
				slog.String("key", key),
				slog.Any("handlerName", config.KeyHandler))
			keymap[key] = config.KeyHandler
		}
	}
	m.processListViewKeymap = keymap
	slog.Info("Initialized process list view keymap",
		slog.Any("handler", m.processListViewKeymap))
}
