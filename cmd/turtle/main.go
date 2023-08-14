package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/alecthomas/kong"

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

	err := cliCtx.Run()
	cliCtx.FatalIfErrorf(err)
}
