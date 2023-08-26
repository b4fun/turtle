package turtle

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

func ExampleSlowloris() {
	u, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	s := Slowloris{
		Target: Target{
			Url: *u,
		},
		SendGibberish: true,
		GibberishInterval: 5 * time.Millisecond,
		UserAgents: []string{
			"turtle/0.0.1",
			"turtle/0.0.1 - slowloris",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = s.Run(ctx)
}

func ExampleSlowBodyReadRequest() {
	u, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	s := SlowBodyReadRequest{
		Target: Target{
			Url: *u,
		},
		Method: http.MethodPost,
		BodyReadTimeout: 180 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = s.Run(ctx)
}