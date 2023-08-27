package main

import (
	"context"
	"os"
	"os/signal"

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
		CLI.Slowloris.Target,
		prog,
	)
	CLI.SlowBodyRead.Target.EventHandler = UIEventHandler(
		CLI.SlowBodyRead.Target,
		prog,
	)

	go func() {
		err := cliCtx.Run()
		prog.Quit()
		if err != nil {
			cliCtx.FatalIfErrorf(err)
		}
	}()

	if _, err := prog.Run(); err != nil {
		cliCtx.FatalIfErrorf(err)
	}
}
