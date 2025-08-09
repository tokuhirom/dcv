package ui

import (
	"log/slog"
	"reflect"
	"runtime"
	"strings"
)

// CommandHandler represents a function that can be executed via command line
type CommandHandler struct {
	Handler     KeyHandler
	Description string
	// TODO: Following comment is completely wrong, ViewMask is not a bitwise mask.
	ViewMask ViewType // Bitwise mask of views where this command is available (0 = all views)
}

// initCommandRegistry initializes the command registry with all available commands
func (m *Model) initCommandRegistry() {
	m.commandRegistry = make(map[string]CommandHandler)

	// Register all key handlers as commands
	m.registerCommands()
}

// registerCommands registers all key handler functions as commands
func (m *Model) registerCommands() {
	// Get all handlers from all views
	allHandlers := []struct {
		handlers []KeyConfig
		viewMask ViewType
	}{
		{m.composeProcessListViewHandlers, ComposeProcessListView},
		{m.logViewHandlers, LogView},
		{m.dindListViewHandlers, DindProcessListView},
		{m.topViewHandlers, TopView},
		{m.statsViewHandlers, StatsView},
		{m.composeProjectListViewHandlers, ComposeProjectListView},
		{m.dockerContainerListViewHandlers, DockerContainerListView},
		{m.imageListViewHandlers, ImageListView},
		{m.networkListViewHandlers, NetworkListView},
		{m.volumeListViewHandlers, VolumeListView},
		{m.fileBrowserHandlers, FileBrowserView},
		{m.fileContentHandlers, FileContentView},
		{m.inspectViewHandlers, InspectView},
		{m.helpViewHandlers, HelpView},
	}

	// Track which handlers we've already registered to avoid duplicates
	registered := make(map[uintptr]bool)

	for _, viewHandlers := range allHandlers {
		for _, handler := range viewHandlers.handlers {
			// Get the function pointer
			funcPtr := reflect.ValueOf(handler.KeyHandler).Pointer()

			// Skip if already registered
			if registered[funcPtr] {
				continue
			}
			registered[funcPtr] = true

			cmdName := GetCommandNameFromFuncPtr(funcPtr)
			slog.Info("Registering command",
				slog.String("cmd", cmdName))
			if cmdName != "" {
				m.commandRegistry[cmdName] = CommandHandler{
					Handler:     handler.KeyHandler,
					Description: handler.Description,
					ViewMask:    0, // XXX
				}
			}
		}
	}
}

func GetCommandNameFromFuncPtr(funcPtr uintptr) string {
	// Get the function name from the pointer
	funcName := runtime.FuncForPC(funcPtr).Name()
	// Extract just the method name (e.g., "SelectUpContainer" from "github.com/tokuhirom/dcv/internal/ui.(*Model).CmdUp-fm")
	parts := strings.Split(funcName, ".")
	if len(parts) == 0 {
		return ""
	}

	methodName := parts[len(parts)-1]
	// Remove the -fm suffix if present (from method value)
	methodName = strings.TrimSuffix(methodName, "-fm")
	// Remove the (*Model) part if present
	methodName = strings.TrimPrefix(methodName, "(*Model)")
	methodName = strings.TrimPrefix(methodName, ")")
	methodName = strings.TrimPrefix(methodName, "Cmd")

	return toKebabCase(methodName)
}

// getCommandForHandler returns the command name for a given key handler
func (m *Model) getCommandForHandler(handler KeyHandler) string {
	if handler == nil {
		return ""
	}

	funcPtr := reflect.ValueOf(handler).Pointer()
	cmdName := GetCommandNameFromFuncPtr(funcPtr)
	return cmdName
}

// toKebabCase converts a CamelCase string to kebab-case
func toKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// getAvailableCommands returns a list of commands available in the current view
func (m *Model) getAvailableCommands() []string {
	var commands []string
	for cmdName, cmd := range m.commandRegistry {
		if cmd.ViewMask == 0 || cmd.ViewMask == m.currentView {
			commands = append(commands, cmdName)
		}
	}
	return commands
}

// getCommandSuggestions returns command suggestions based on partial input
func (m *Model) getCommandSuggestions(partial string) []string {
	var suggestions []string
	availableCommands := m.getAvailableCommands()

	for _, cmdName := range availableCommands {
		if strings.HasPrefix(cmdName, partial) {
			suggestions = append(suggestions, cmdName)
		}
	}

	// If no prefix matches, try substring matching
	if len(suggestions) == 0 {
		for _, cmdName := range availableCommands {
			if strings.Contains(cmdName, partial) {
				suggestions = append(suggestions, cmdName)
			}
		}
	}

	return suggestions
}
