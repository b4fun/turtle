package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/b4fun/turtle"
)

var CLI struct {
	Slowloris    turtle.Slowloris           `cmd:"slowloris" help:"Run slowloris attack"`
	SlowBodyRead turtle.SlowBodyReadRequest `cmd:"slow-body-read" help:"Run slow body read attack"`
}

func main() {
	cliCtx := kong.Parse(&CLI)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cliCtx.BindTo(ctx, (*context.Context)(nil))

	prog := tea.NewProgram(newUI())
	CLI.Slowloris.Target.EventHandler = UIEventHandler(
		ctx,
		prog,
		CLI.Slowloris.Target,
		500*time.Millisecond,
	)
	CLI.SlowBodyRead.Target.EventHandler = UIEventHandler(
		ctx,
		prog,
		CLI.SlowBodyRead.Target,
		500*time.Millisecond,
	)

	go func() {
		if _, err := prog.Run(); err != nil {
			cliCtx.FatalIfErrorf(err)
		}
		cancel()
	}()

	err := cliCtx.Run()
	go prog.Quit()
	if err != nil {
		cliCtx.FatalIfErrorf(err)
	}
}
