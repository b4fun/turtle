package testhttp

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func newLocalListener() (net.Listener, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			return nil, err
		}
	}
	return l, nil
}

func CreateTestServer(ctx context.Context, t testing.TB, mu ...func(s *http.Server)) (*url.URL, <-chan struct{}) {
	t.Helper()

	ln, err := newLocalListener()
	require.NoError(t, err)
	t.Logf("listening at %q", ln.Addr())

	t.Cleanup(func() {
		_ = ln.Close()
	})

	server := &http.Server{
		Addr: ln.Addr().String(),
	}
	for _, m := range mu {
		m(server)
	}

	serverStopped := make(chan struct{})
	go func() {
		_ = server.Serve(ln)
		close(serverStopped)
	}()
	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	u, err := url.Parse("http://" + ln.Addr().String())
	require.NoError(t, err)

	return u, serverStopped
}

type ConnStateCounters struct {
	mu sync.Mutex // protects below	
	connTimeline []string // value is connection's remote addr
	connStatesTimeline map[string][]http.ConnState // key is connection's remote addr
}

func NewConnStateCounters() *ConnStateCounters {
	return &ConnStateCounters{
		connStatesTimeline: make(map[string][]http.ConnState),
	}
}

func (c *ConnStateCounters) ServerConnState(conn net.Conn, state http.ConnState) {
	remoteAddr := conn.RemoteAddr().String()

	c.mu.Lock()
	defer c.mu.Unlock()

	statesTimeline, exists := c.connStatesTimeline[remoteAddr]
	if !exists {
		c.connTimeline = append(c.connTimeline, remoteAddr)
	}
	c.connStatesTimeline[remoteAddr] = append(statesTimeline, state)
}

func (c *ConnStateCounters) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	var s string
	for _, addr := range c.connTimeline {
		s += addr + ":"
		timeline := c.connStatesTimeline[addr]
		if len(timeline) == 0 {
			s += " (no states)\n"
			continue
		}
		for idx, state := range timeline {
			s += state.String()
			if idx < len(timeline) - 1 {
				s += " -> "
			}
		}
		s += "\n"
	}
	return s
}

func (c *ConnStateCounters) GetConns() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.connTimeline[:]
}

func (c *ConnStateCounters) GetConnStateTimeline(addr string) []http.ConnState {
	c.mu.Lock()
	defer c.mu.Unlock()

	v, exists := c.connStatesTimeline[addr]
	if !exists {
		return nil
	}

	return v[:]
}