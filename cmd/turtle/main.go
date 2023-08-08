package main

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/b4fun/turtle"
)

func main() {
	s := turtle.Slowloris{
		Target: &turtle.Target{
			Url: func() url.URL {
				p, err := url.Parse("http://localhost:3333")
				if err != nil {
					panic(err)
				}
				return *p
			}(),
			Duration:   30 * time.Second,
			Connections: 100,
		},
		SendGibberish: true,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := s.Run(ctx); err != nil {
		panic(err)
	}
}