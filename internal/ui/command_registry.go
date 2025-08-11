package ui

import (
	"reflect"
	"runtime"
	"strings"
)

// CommandHandler represents a function that can be executed via command line
type CommandHandler struct {
	Handler     KeyHandler
	Description string
}

// initCommandRegistry initializes the command registry with all available commands
func (m *Model) initCommandRegistry() {
	m.allCommands = make(map[string]struct{})
	m.viewCommandRegistry = make(map[ViewType]map[string]CommandHandler)

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

	for _, viewHandlers := range allHandlers {
		m.viewCommandRegistry[viewHandlers.viewMask] = make(map[string]CommandHandler)

		for _, handler := range viewHandlers.handlers {
			// Get the function pointer
			funcPtr := reflect.ValueOf(handler.KeyHandler).Pointer()

			cmdName := GetCommandNameFromFuncPtr(funcPtr)
			if cmdName != "" {
				m.viewCommandRegistry[viewHandlers.viewMask][cmdName] = CommandHandler{
					Handler:     handler.KeyHandler,
					Description: handler.Description,
				}

				m.allCommands[cmdName] = struct{}{}
			}
		}
	}

	// Register global handlers (they work across all views)
	for _, handler := range m.globalHandlers {
		funcPtr := reflect.ValueOf(handler.KeyHandler).Pointer()
		cmdName := GetCommandNameFromFuncPtr(funcPtr)
		if cmdName != "" {
			m.allCommands[cmdName] = struct{}{}
			// Add to all view registries since global commands work everywhere
			for viewType := range m.viewCommandRegistry {
				m.viewCommandRegistry[viewType][cmdName] = CommandHandler{
					Handler:     handler.KeyHandler,
					Description: handler.Description,
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
	// Handle special cases for common abbreviations
	switch s {
	case "PS":
		return "ps"
	case "ComposeLS":
		return "compose-ls"
	}

	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
