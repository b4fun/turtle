package turtle

import (
	"net/url"
	"time"
)

// Target provides target configuration.
type Target struct {
	// Url - the url of the target.
	Url url.URL
	// Duration - the duration of the attack. Defaults to 30s.
	Duration time.Duration
	// Connections - the number of connections to be made. Defaults to 100.
	Connections int
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