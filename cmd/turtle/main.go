package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/sparkline"

	"github.com/b4fun/turtle"
)

var CLI struct {
	Slowloris    turtle.Slowloris           `cmd:"slowloris" help:"Run slowloris attack"`
	SlowBodyRead turtle.SlowBodyReadRequest `cmd:"slow-body-read" help:"Run slow body read attack"`
}

type SparkLineEventHandler struct {
	refreshRate time.Duration

	mu          sync.Mutex
	dial        *sparkline.SparkLine
	closed      *sparkline.SparkLine
	countDial   int
	countClosed int
}

func NewSparkLineEventHandler() (*SparkLineEventHandler, error) {
	dial, err := sparkline.New(
		sparkline.Label("#Dial", cell.FgColor(cell.ColorNumber(33))),
		sparkline.Color(cell.ColorGreen),
	)
	if err != nil {
		return nil, err
	}
	closed, err := sparkline.New(
		sparkline.Label("#Closed", cell.FgColor(cell.ColorNumber(33))),
		sparkline.Color(cell.ColorRed),
	)
	if err != nil {
		return nil, err
	}

	rv := &SparkLineEventHandler{
		refreshRate: time.Second,

		dial:   dial,
		closed: closed,
	}

	return rv, nil
}

func (s *SparkLineEventHandler) tick() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := errors.Join(
		s.dial.Add([]int{s.countDial}),
		s.closed.Add([]int{s.countClosed}),
	); err != nil {
		return err
	}

	s.countDial = 0
	s.countClosed = 0

	return nil
}

func (s *SparkLineEventHandler) runTicker(ctx context.Context) error {
	ticker := time.NewTicker(s.refreshRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.tick(); err != nil {
				return err
			}
		}
	}
}

func (s *SparkLineEventHandler) HandleEvent(e turtle.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch e.Name {
	case turtle.EventTCPDial:
		s.countDial++
	case turtle.EventTCPClosed:
		s.countClosed++
	}
}

func (s *SparkLineEventHandler) drawDashboard(ctx context.Context, cancel context.CancelFunc) error {
	term, err := tcell.New()
	if err != nil {
		return err
	}

	c, err := container.New(
		term,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.Border(linestyle.Light),
		container.BorderTitle("Network Connections"),
		container.SplitHorizontal(
			container.Top(
				container.PlaceWidget(s.dial),
			),
			container.Bottom(
				container.PlaceWidget(s.closed),
			),
		),
	)
	if err != nil {
		return err
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	return termdash.Run(ctx, term, c, termdash.KeyboardSubscriber(quitter))
}

func (s *SparkLineEventHandler) Start(cliCtx *kong.Context) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cliCtx.BindTo(ctx, (*context.Context)(nil))
	errChan := make(chan error, 3)

	go func() {
		if err := s.runTicker(ctx); err != nil {
			errChan <- err
		}
	}()

	go func() {
		if err := s.drawDashboard(ctx, cancel); err != nil {
			errChan <- err
		}
	}()

	go func() {
		if err := cliCtx.Run(); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		return
	case err := <-errChan:
		cliCtx.FatalIfErrorf(err)
	}
}

func main() {
	cliCtx := kong.Parse(&CLI)
	sl, err := NewSparkLineEventHandler()
	if err != nil {
		cliCtx.FatalIfErrorf(err)
	}
	eventHandler := turtle.NewAsyncEventHandler(sl, 1000)

	CLI.Slowloris.Target.EventHandler = eventHandler
	CLI.SlowBodyRead.Target.EventHandler = eventHandler

	sl.Start(cliCtx)
}
