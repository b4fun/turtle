package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/b4fun/turtle"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UIUpdate struct {
	target         string
	countConnected int64
	countClosed    int64
	eventDescs     []string
}

func formatEvent(event turtle.Event) string {
	var s string

	err, hasErr := turtle.ErrorFromEvent(event)
	if hasErr {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("9")).SetString("x").String()
	} else {
		s += " "
	}

	workerId, hasWorkerId := turtle.WorkerIdFromEvent(event)
	if hasWorkerId {
		s += fmt.Sprintf(" [%03d]", workerId)
	}

	s += " " + event.Name

	if hasErr {
		s += " " + err.Error()
	}

	return s
}

func UIEventHandler(
	ctx context.Context,
	program *tea.Program,
	target turtle.Target,
	updateInterval time.Duration,
) turtle.EventHandler {
	var (
		startOnce sync.Once
		l         sync.Mutex
		update    UIUpdate
	)

	start := func() {
		go func() {
			update.target = target.Url.String()

			ticker := time.NewTicker(updateInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					var u UIUpdate
					l.Lock()
					u = update
					update.eventDescs = nil
					l.Unlock()

					program.Send(u)
				}
			}
		}()
	}

	h := turtle.EventHandleFunc(func(event turtle.Event) {
		startOnce.Do(start)

		l.Lock()
		switch event.Name {
		case turtle.EventTCPDial:
			update.countConnected += 1
		case turtle.EventTCPClosed:
			update.countClosed += 1
		case turtle.EventWorkerError, turtle.EventWorkerPanic:
			update.eventDescs = append(
				update.eventDescs,
				formatEvent(event),
			)
		}
		l.Unlock()
	})

	return turtle.NewAsyncEventHandler(h, 1000)
}

type UI struct {
	width  int
	height int

	target         string
	countConnected int64
	countClosed    int64

	spinner  spinner.Model
	progress progress.Model
}

func newUI() *UI {
	s := spinner.New(spinner.WithSpinner(spinner.Meter))
	p := progress.New(
		progress.WithSolidFill("#5A56E0"),
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

		ui.target = message.target
		ui.countConnected = message.countConnected
		ui.countClosed = message.countClosed
		ratio := float64(message.countConnected) / float64(message.countConnected+message.countClosed)

		// Update progress bar
		cmds = append(
			cmds,
			ui.progress.SetPercent(ratio),
		)

		// Print event log
		if len(message.eventDescs) > 0 {
			cmds = append(
				cmds,
				tea.Printf(strings.Join(message.eventDescs, "\n")),
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
	bannerLine := ui.spinner.View() + " (press q to quit)"
	targetLine := fmt.Sprintf("target: %s", ui.target)

	prog := ui.progress.View()
	cellsAvail := max(0, ui.width-lipgloss.Width(prog))

	infoDesc := fmt.Sprintf("total connected %d / total closed %d ", ui.countConnected, ui.countClosed)
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render(infoDesc)

	return "\n" + targetLine + "\n" + info + prog + "\n" + bannerLine
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
