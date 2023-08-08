package turtle

import (
	"net/url"
	"time"
)

// Target provides target configuration.
type Target struct {
	// Url - the url of the target.
	Url url.URL `arg:"" name:"target-url" help:"the url of the target"`
	// Duration - the duration of the attack. Defaults to 30s.
	Duration time.Duration `name:"target-duration" help:"the duration of the attack. Defaults to 30s"`
	// Connections - the number of connections to be made. Defaults to 100.
	Connections int `name:"target-connections" help:"the number of connections to be made. Defaults to 100"`
}

func (t *Target) defaults() error {
	if t.Duration <= 0 {
		t.Duration = 30 * time.Second
	}
	if t.Connections < 1 {
		t.Connections = 100
	}

	return nil
}