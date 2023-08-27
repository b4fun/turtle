package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/b4fun/turtle"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UIUpdate struct {
	CountConnected int
	CountTotal     int
	EventDesc      string
}

func formatEvent(event turtle.Event) string {
	var s string

	err, hasErr := turtle.ErrorFromEvent(event)
	if hasErr {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("9")).SetString("x").String()
	}

	workerId, hasWorkerId := turtle.WorkerIdFromEvent(event)
	if hasWorkerId {
		s += fmt.Sprintf(" [%s]", workerId)
	}

	s += " " + event.Name

	if hasErr {
		s += " " + err.Error()
	}

	return s
}

func UIEventHandler(target turtle.Target, program *tea.Program) turtle.EventHandler {
	countTotal := target.Connections

	var (
		l           sync.Mutex
		countDialed int
	)

	h := turtle.EventHandleFunc(func(event turtle.Event) {
		var update UIUpdate
		update.CountTotal = countTotal

		l.Lock()
		switch event.Name {
		case turtle.EventTCPDial:
			countDialed = max(countTotal, countDialed+1)
			update.EventDesc = formatEvent(event)
		case turtle.EventTCPClosed:
			countDialed = max(0, countDialed-1)
			update.EventDesc = formatEvent(event)
		case turtle.EventWorkerError, turtle.EventWorkerPanic:
			update.EventDesc = formatEvent(event)
		}
		l.Unlock()

		update.CountConnected = countDialed

		program.Send(update)
	})

	return turtle.NewAsyncEventHandler(h, 1000)
}

type UI struct {
	width  int
	height int

	spinner  spinner.Model
	progress progress.Model
}

func newUI() *UI {
	s := spinner.New(spinner.WithSpinner(spinner.Meter))
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return &UI{
		spinner:  s,
		progress: p,
	}
}

var _ tea.Model = &UI{}

func (ui *UI) Init() tea.Cmd {
	return tea.Batch(ui.spinner.Tick)
}

func (ui *UI) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		ui.width, ui.height = message.Width, message.Height
	case tea.KeyMsg:
		switch message.String() {
		case "ctrl+c", "esc", "q":
			return ui, tea.Quit
		}
	case UIUpdate:
		var cmds []tea.Cmd

		// Update progress bar
		cmds = append(
			cmds,
			ui.progress.SetPercent(float64(message.CountConnected)/float64(message.CountTotal)),
		)

		// Print event log
		if message.EventDesc != "" {
			cmds = append(
				cmds,
				tea.Printf(message.EventDesc),
			)
		}

		return ui, tea.Batch(cmds...)
	case spinner.TickMsg:
		var cmd tea.Cmd
		ui.spinner, cmd = ui.spinner.Update(message)
		return ui, cmd
	case progress.FrameMsg:
		newModel, cmd := ui.progress.Update(message)
		if newModel, ok := newModel.(progress.Model); ok {
			ui.progress = newModel
		}
		return ui, cmd
	}

	return ui, nil
}

func (ui *UI) View() string {
	spin := ui.spinner.View() + " "
	prog := ui.progress.View()

	cellsRemaining := max(0, ui.width-lipgloss.Width(spin+prog))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + prog + gap
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
