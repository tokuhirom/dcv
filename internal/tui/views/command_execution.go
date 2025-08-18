package views

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CommandExecutionView displays real-time output from Docker command execution
type CommandExecutionView struct {
	textView     *tview.TextView
	flex         *tview.Flex
	pages        *tview.Pages
	command      string
	cmd          *exec.Cmd
	output       []string
	outputMux    sync.Mutex
	done         bool
	exitCode     int
	onClose      func()
	autoScroll   bool
	maxLines     int
	currentLines int
}

// NewCommandExecutionView creates a new command execution view
func NewCommandExecutionView() *CommandExecutionView {
	v := &CommandExecutionView{
		textView:   tview.NewTextView(),
		pages:      tview.NewPages(),
		output:     make([]string, 0),
		autoScroll: true,
		maxLines:   10000, // Limit output to prevent memory issues
	}

	v.setupView()
	return v
}

// setupView configures the view components
func (v *CommandExecutionView) setupView() {
	// Configure text view for output
	v.textView.
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			if v.autoScroll {
				v.textView.ScrollToEnd()
			}
		})

	// Create a flex layout for the modal
	v.flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle(" Command Execution "), 0, 6, false).
			AddItem(nil, 0, 1, false), 0, 8, false).
		AddItem(nil, 0, 1, false)

	// Configure the inner content box
	innerFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(v.createHeader(), 3, 0, false).
		AddItem(v.textView, 0, 1, true).
		AddItem(v.createFooter(), 1, 0, false)

	// Update the middle box with content
	v.flex.RemoveItem(v.flex.GetItem(1))
	v.flex.AddItem(tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(innerFlex.SetBorder(true).SetTitle(" Command Execution "), 0, 6, true).
		AddItem(nil, 0, 1, false), 0, 8, true)

	// Set up keyboard handlers
	v.setupKeyHandlers()

	// Add to pages
	v.pages.AddPage("modal", v.flex, true, true)
}

// createHeader creates the header showing the command
func (v *CommandExecutionView) createHeader() tview.Primitive {
	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	header.SetText(fmt.Sprintf("[yellow]Command:[-] %s\n", v.command))
	return header
}

// createFooter creates the footer with instructions
func (v *CommandExecutionView) createFooter() tview.Primitive {
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	if v.done {
		if v.exitCode == 0 {
			footer.SetText("[green]✓ Command completed successfully[-] | Press [yellow]ESC[-] or [yellow]q[-] to close")
		} else {
			footer.SetText(fmt.Sprintf("[red]✗ Command failed with exit code %d[-] | Press [yellow]ESC[-] or [yellow]q[-] to close", v.exitCode))
		}
	} else {
		footer.SetText("[yellow]⠋ Running...[-] | Press [yellow]Ctrl+C[-] to cancel | [yellow]ESC[-] to close")
	}

	return footer
}

// setupKeyHandlers sets up keyboard shortcuts
func (v *CommandExecutionView) setupKeyHandlers() {
	v.textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			v.close()
			return nil
		case tcell.KeyCtrlC:
			if !v.done && v.cmd != nil {
				v.cancelCommand()
			}
			return nil
		case tcell.KeyPgUp:
			v.autoScroll = false
			// Let default handler process
			return event
		case tcell.KeyPgDn:
			// Let default handler process
			return event
		case tcell.KeyEnd:
			v.autoScroll = true
			v.textView.ScrollToEnd()
			return nil
		case tcell.KeyHome:
			v.autoScroll = false
			v.textView.ScrollToBeginning()
			return nil
		}

		switch event.Rune() {
		case 'q', 'Q':
			v.close()
			return nil
		case 'k':
			v.autoScroll = false
			row, col := v.textView.GetScrollOffset()
			if row > 0 {
				v.textView.ScrollTo(row-1, col)
			}
			return nil
		case 'j':
			row, col := v.textView.GetScrollOffset()
			v.textView.ScrollTo(row+1, col)
			return nil
		case 'g':
			v.autoScroll = false
			v.textView.ScrollToBeginning()
			return nil
		case 'G':
			v.autoScroll = true
			v.textView.ScrollToEnd()
			return nil
		}

		return event
	})
}

// GetPrimitive returns the tview primitive for this view
func (v *CommandExecutionView) GetPrimitive() tview.Primitive {
	return v.pages
}

// ExecuteCommand executes a Docker command and displays the output
func (v *CommandExecutionView) ExecuteCommand(command string, args ...string) {
	v.command = fmt.Sprintf("docker %s", strings.Join(args, " "))
	v.output = make([]string, 0)
	v.done = false
	v.exitCode = 0
	v.currentLines = 0
	v.autoScroll = true

	// Clear previous output
	v.textView.Clear()

	// Update header with new command
	v.updateDisplay()

	// Start command execution in background
	go v.runCommand(args)
}

// ExecuteDockerCommand is a convenience method for executing docker commands
func (v *CommandExecutionView) ExecuteDockerCommand(args ...string) {
	v.ExecuteCommand("docker", args...)
}

// runCommand executes the command and streams output
func (v *CommandExecutionView) runCommand(args []string) {
	slog.Info("Executing command in modal", slog.String("command", v.command))

	// Create the command
	v.cmd = exec.Command("docker", args...)

	// Create pipes for stdout and stderr
	stdout, err := v.cmd.StdoutPipe()
	if err != nil {
		v.appendOutput(fmt.Sprintf("[red]Error creating stdout pipe: %v[-]", err))
		v.done = true
		v.exitCode = 1
		QueueUpdateDraw(func() {
			v.updateDisplay()
		})
		return
	}

	stderr, err := v.cmd.StderrPipe()
	if err != nil {
		v.appendOutput(fmt.Sprintf("[red]Error creating stderr pipe: %v[-]", err))
		v.done = true
		v.exitCode = 1
		QueueUpdateDraw(func() {
			v.updateDisplay()
		})
		return
	}

	// Start the command
	if err := v.cmd.Start(); err != nil {
		v.appendOutput(fmt.Sprintf("[red]Error starting command: %v[-]", err))
		v.done = true
		v.exitCode = 1
		QueueUpdateDraw(func() {
			v.updateDisplay()
		})
		return
	}

	// Create a multi-reader to merge stdout and stderr
	multiReader := io.MultiReader(stdout, stderr)
	scanner := bufio.NewScanner(multiReader)

	// Read output line by line
	for scanner.Scan() {
		line := scanner.Text()
		v.appendOutput(line)

		// Update display periodically
		QueueUpdateDraw(func() {
			v.updateDisplay()
		})
	}

	// Wait for command to complete
	if err := v.cmd.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			v.exitCode = exitError.ExitCode()
		} else {
			v.exitCode = 1
		}
	} else {
		v.exitCode = 0
	}

	v.done = true
	slog.Info("Command completed",
		slog.String("command", v.command),
		slog.Int("exitCode", v.exitCode))

	// Final update
	QueueUpdateDraw(func() {
		v.updateDisplay()
	})
}

// appendOutput adds a line to the output buffer
func (v *CommandExecutionView) appendOutput(line string) {
	v.outputMux.Lock()
	defer v.outputMux.Unlock()

	// Limit the number of lines to prevent memory issues
	if v.currentLines >= v.maxLines {
		// Remove oldest lines
		if len(v.output) > 100 {
			v.output = v.output[100:]
			v.currentLines -= 100
		}
	}

	v.output = append(v.output, line)
	v.currentLines++
}

// updateDisplay updates the text view with current output
func (v *CommandExecutionView) updateDisplay() {
	v.outputMux.Lock()
	outputText := strings.Join(v.output, "\n")
	v.outputMux.Unlock()

	v.textView.SetText(outputText)

	// Update footer based on status if flex is properly initialized
	if v.flex != nil && v.flex.GetItemCount() > 1 {
		item := v.flex.GetItem(1)
		if flexItem, ok := item.(*tview.Flex); ok && flexItem.GetItemCount() > 1 {
			innerItem := flexItem.GetItem(1)
			if innerFlex, ok := innerItem.(*tview.Flex); ok && innerFlex.GetItemCount() > 0 {
				footerIndex := innerFlex.GetItemCount() - 1
				innerFlex.RemoveItem(innerFlex.GetItem(footerIndex))
				innerFlex.AddItem(v.createFooter(), 1, 0, false)
			}
		}
	}
}

// cancelCommand cancels the running command
func (v *CommandExecutionView) cancelCommand() {
	if v.cmd != nil && v.cmd.Process != nil && !v.done {
		slog.Info("Cancelling command", slog.String("command", v.command))

		// Kill the process
		if err := v.cmd.Process.Kill(); err != nil {
			v.appendOutput(fmt.Sprintf("[red]Error cancelling command: %v[-]", err))
		} else {
			v.appendOutput("[yellow]Command cancelled by user[-]")
		}

		v.done = true
		v.exitCode = -1
		v.updateDisplay()
	}
}

// close closes the modal and returns to the previous view
func (v *CommandExecutionView) close() {
	// Cancel command if still running
	if !v.done && v.cmd != nil {
		v.cancelCommand()
	}

	// Call the close callback if set
	if v.onClose != nil {
		v.onClose()
	}
}

// SetOnClose sets the callback to be called when the modal is closed
func (v *CommandExecutionView) SetOnClose(fn func()) {
	v.onClose = fn
}

// IsRunning returns true if a command is currently executing
func (v *CommandExecutionView) IsRunning() bool {
	return !v.done
}

// GetExitCode returns the exit code of the last command
func (v *CommandExecutionView) GetExitCode() int {
	return v.exitCode
}

// ShowConfirmationAndExecute shows a confirmation dialog before executing an aggressive command
func (v *CommandExecutionView) ShowConfirmationAndExecute(command string, args []string, onConfirm func()) {
	commandStr := fmt.Sprintf("docker %s", strings.Join(args, " "))
	text := fmt.Sprintf("Are you sure you want to execute:\n\n%s", commandStr)

	modal := CreateConfirmationModal(text,
		func() {
			// User confirmed, execute the command
			v.ExecuteDockerCommand(args...)
			if onConfirm != nil {
				onConfirm()
			}
		},
		func() {
			// User cancelled, close the modal
			if v.onClose != nil {
				v.onClose()
			}
		})

	// Show confirmation modal
	v.pages.AddPage("confirm", modal, true, true)
}

// ExecuteWithProgress executes a command and shows progress in the modal
func ExecuteWithProgress(args []string, onClose func()) *CommandExecutionView {
	view := NewCommandExecutionView()
	view.SetOnClose(onClose)
	view.ExecuteDockerCommand(args...)
	return view
}

// ExecuteAggressiveCommand executes a command that requires confirmation
func ExecuteAggressiveCommand(args []string, onClose func()) *CommandExecutionView {
	view := NewCommandExecutionView()
	view.SetOnClose(onClose)
	view.ShowConfirmationAndExecute("docker", args, nil)
	return view
}

// GetTitle returns the title of the view (implements View interface)
func (v *CommandExecutionView) GetTitle() string {
	return "Command Execution"
}

// Refresh refreshes the view (implements View interface)
func (v *CommandExecutionView) Refresh() {
	// Command execution doesn't need refresh - it's a one-time operation
}
