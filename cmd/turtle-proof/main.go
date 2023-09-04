package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/alecthomas/kong"
)

var logger = slog.Default().WithGroup("turtle-proof")

type CLI struct {
	ServerAddr string `cmd:"server-addr" help:"the address to listen on" default:"127.0.0.1:8889"`
	Scenario   string `cmd:"scenario" enum:"none,proof" help:"the scenario to run" default:"none"`

	ServerReadHeaderTimeout time.Duration `cmd:"server-read-header-timeout" help:"the amount of time allowed to read request headers. Non-positive value means no timeout"`
	ServerReadTimeout       time.Duration `cmd:"server-read-timeout" help:"the maximum duration for reading the entire request, including the body. Non-positive value means no timeout"`
	ServerWriteTimeout      time.Duration `cmd:"server-write-timeout" help:"the maximum duration before timing out writes of the response. Non-positive value means no timeout"`
}

func (c *CLI) defaults() error {
	switch c.Scenario {
	case "proof":
		if c.ServerReadHeaderTimeout <= 0 {
			c.ServerReadHeaderTimeout = 3 * time.Second
		}
		if c.ServerReadTimeout <= 0 {
			c.ServerReadTimeout = 60 * time.Second
		}
		if c.ServerWriteTimeout <= 0 {
			c.ServerWriteTimeout = 60 * time.Second
		}
	default:
		if c.ServerReadHeaderTimeout <= 0 {
			c.ServerReadHeaderTimeout = 0
		}
		if c.ServerReadTimeout <= 0 {
			c.ServerReadTimeout = 0
		}
		if c.ServerWriteTimeout <= 0 {
			c.ServerWriteTimeout = 0
		}
	}

	return nil
}

func (c *CLI) CreateServer() *http.Server {
	return &http.Server{
		Addr:              c.ServerAddr,
		ReadHeaderTimeout: c.ServerReadHeaderTimeout,
		ReadTimeout:       c.ServerReadTimeout,
		WriteTimeout:      c.ServerWriteTimeout,

		ConnState: func(conn net.Conn, state http.ConnState) {
			logger.Info(
				"ConnState",
				slog.String("remote_addr", conn.RemoteAddr().String()),
				slog.String("state", state.String()),
			)
		},

		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("handling request")

			if r.Body != nil {
				logger.Info("reading body")
				if _, err := io.ReadAll(r.Body); err != nil {
					logger.Error("read error", slog.String("error", err.Error()))
					w.WriteHeader(http.StatusInternalServerError)
				}
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}),
	}
}

const description = `
Proof HTTP server for turtle test.

## Start with default settings

	turtle-proof

## Start with proofed settings

	turtle-proof --scenario proof

## Using custom settings

	turtle-proof --server-addr="127.0.0.1:8888" --server-read-header-timeout 3s --server-read-timeout 60s --server-write-timeout 60s
`

func main() {
	cli := &CLI{}
	cliCtx := kong.Parse(
		cli,
		kong.Name("turtle-proof"),
		kong.Description(description),
	)

	if err := cli.defaults(); err != nil {
		cliCtx.FatalIfErrorf(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	server := cli.CreateServer()
	go func() {
		err := server.ListenAndServe()
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			cliCtx.FatalIfErrorf(err)
		}
	}()

	time.Sleep(300 * time.Millisecond) // wait for server start
	logger.Info("server started", slog.String("addr", server.Addr))

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(shutdownCtx)
}
