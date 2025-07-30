package ui

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type LogView struct {
	app           *App
	view          *tview.TextView
	containerName string
	isDind        bool
	hostContainer string // For dind logs
	searchMode    bool
	searchText    string
}

func NewLogView(app *App) *LogView {
	v := &LogView{
		app:  app,
		view: tview.NewTextView(),
	}

	v.setupView()
	v.setupKeyBindings()

	return v
}

func (v *LogView) setupView() {
	v.view.SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetScrollable(true).
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 1)
}

func (v *LogView) setupKeyBindings() {
	v.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if v.searchMode {
			return v.handleSearchMode(event)
		}

		switch event.Key() {
		case tcell.KeyEsc:
			v.app.ShowProcessList()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				v.app.ShowProcessList()
				return nil
			case 'G':
				v.view.ScrollToEnd()
				return nil
			case 'g':
				v.view.ScrollToBeginning()
				return nil
			case '/':
				v.enterSearchMode()
				return nil
			case 'n':
				v.findNext()
				return nil
			case 'N':
				v.findPrevious()
				return nil
			case 'j':
				row, col := v.view.GetScrollOffset()
				v.view.ScrollTo(row+1, col)
				return nil
			case 'k':
				row, col := v.view.GetScrollOffset()
				if row > 0 {
					v.view.ScrollTo(row-1, col)
				}
				return nil
			}
		}
		return event
	})
}

func (v *LogView) SetContainer(containerName string, isDind bool) {
	v.containerName = containerName
	v.isDind = isDind
	v.hostContainer = ""
	v.view.Clear()
	v.view.SetTitle(fmt.Sprintf(" Logs: %s ", containerName))
	
	go v.streamLogs()
}

func (v *LogView) SetDindContainer(hostContainer, targetContainer string) {
	v.hostContainer = hostContainer
	v.containerName = targetContainer
	v.isDind = true
	v.view.Clear()
	v.view.SetTitle(fmt.Sprintf(" Logs: %s (in %s) ", targetContainer, hostContainer))
	
	go v.streamLogs()
}

func (v *LogView) streamLogs() {
	var cmd *exec.Cmd
	var err error
	
	if v.isDind && v.hostContainer != "" {
		// Get logs from container inside dind
		cmd, err = v.app.dockerClient.GetDindContainerLogs(v.hostContainer, v.containerName, true)
	} else {
		// Get regular container logs
		cmd, err = v.app.dockerClient.GetContainerLogs(v.containerName, true)
	}
	
	if err != nil {
		v.view.SetText(fmt.Sprintf("Error: %v", err))
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		v.view.SetText(fmt.Sprintf("Error: %v", err))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		v.view.SetText(fmt.Sprintf("Error: %v", err))
		return
	}

	if err := cmd.Start(); err != nil {
		v.view.SetText(fmt.Sprintf("Error starting command: %v", err))
		return
	}

	// Merge stdout and stderr
	go v.readLogs(stdout)
	go v.readLogs(stderr)
}

func (v *LogView) readLogs(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		v.app.app.QueueUpdateDraw(func() {
			fmt.Fprintln(v.view, line)
		})
	}
}

func (v *LogView) enterSearchMode() {
	v.searchMode = true
	v.searchText = ""
	v.updateTitle()
}

func (v *LogView) handleSearchMode(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		v.searchMode = false
		v.updateTitle()
		return nil
	case tcell.KeyEnter:
		v.searchMode = false
		v.findNext()
		v.updateTitle()
		return nil
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(v.searchText) > 0 {
			v.searchText = v.searchText[:len(v.searchText)-1]
			v.updateTitle()
		}
		return nil
	case tcell.KeyRune:
		v.searchText += string(event.Rune())
		v.updateTitle()
		return nil
	}
	return event
}

func (v *LogView) updateTitle() {
	if v.searchMode {
		v.view.SetTitle(fmt.Sprintf(" Search: %s ", v.searchText))
	} else {
		v.view.SetTitle(fmt.Sprintf(" Logs: %s ", v.containerName))
	}
}

func (v *LogView) findNext() {
	if v.searchText == "" {
		return
	}
	
	text := v.view.GetText(true)
	currentPos, _ := v.view.GetScrollOffset()
	
	// Search from current position
	index := strings.Index(text[currentPos:], v.searchText)
	if index >= 0 {
		v.view.ScrollTo(currentPos+index, 0)
		v.view.Highlight(v.searchText)
	}
}

func (v *LogView) findPrevious() {
	if v.searchText == "" {
		return
	}
	
	text := v.view.GetText(true)
	currentPos, _ := v.view.GetScrollOffset()
	
	// Search backward from current position
	if currentPos > 0 {
		index := strings.LastIndex(text[:currentPos], v.searchText)
		if index >= 0 {
			v.view.ScrollTo(index, 0)
			v.view.Highlight(v.searchText)
		}
	}
}