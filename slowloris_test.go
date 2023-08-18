package turtle

import (
	"net/http"
	"testing"
	"time"

	"github.com/b4fun/turtle/internal/testhttp"
	"github.com/stretchr/testify/assert"
)

func Test_Slowloris_defaults(t *testing.T) {
	v := &Slowloris{}
	assert.NoError(t, v.defaults())

	assert.Equal(t, http.MethodGet, v.Method)
	assert.NotEmpty(t, v.UserAgents)
	assert.NotNil(t, v.dial)
	assert.NotNil(t, v.randn)
}

func Test_Slowloris_vulnerable(t *testing.T) {
	t.Parallel()

	counter := testhttp.NewConnStateCounters()
	var serverStopped <-chan struct{}

	testCtx, stopTest := getTestContext()
	defer stopTest()

	serverUrl, serverStopped := testhttp.CreateTestServer(testCtx, t, func(s *http.Server) {
		s.ReadHeaderTimeout = 0 // hung forever
		s.ConnState = counter.ServerConnState
	})

	target := getTestTarget(t, *serverUrl)
	attack := Slowloris{Target: target, SendGibberish: true, GibberishInterval: 10 * time.Millisecond}

	assert.NoError(t, attack.Run(testCtx))
	stopTest()

	t.Log(counter.String())
	<-serverStopped

	conns := counter.GetConns()
	assert.NotEmpty(t, conns, "should have at least one connection")
	assert.Len(t, conns, target.Connections, "should not create more connections")
	for _, conn := range conns {
		timeline := counter.GetConnStateTimeline(conn)
		assert.GreaterOrEqual(t, len(timeline), 1, "should have at least one state")
		// a typical state transition is: new - (due to server close) -> active -> closed
		assert.Equal(t, http.StateNew, timeline[0])
	}
}

func Test_Slowloris_invulnerable(t *testing.T) {
	t.Parallel()

	counter := testhttp.NewConnStateCounters()
	var serverStopped <-chan struct{}

	testCtx, stopTest := getTestContext()
	defer stopTest()

	serverUrl, serverStopped := testhttp.CreateTestServer(testCtx, t, func(s *http.Server) {
		s.ReadHeaderTimeout = 1 * time.Second
		s.ConnState = counter.ServerConnState
	})

	target := getTestTarget(t, *serverUrl)
	attack := Slowloris{Target: target, SendGibberish: true, GibberishInterval: 10 * time.Millisecond}

	assert.NoError(t, attack.Run(testCtx))
	stopTest()

	t.Log(counter.String())
	<-serverStopped

	conns := counter.GetConns()
	assert.NotEmpty(t, conns, "should have at least one connection")
	assert.Greater(t, len(conns), target.Connections, "should create more connections")
	for _, conn := range conns {
		timeline := counter.GetConnStateTimeline(conn)
		assert.GreaterOrEqual(t, len(timeline), 2, "should have at least two state")
		// a typical state transition is: new - (due to read header timeout close) -> active -> closed
		assert.Equal(t, http.StateNew, timeline[0])
	}
}
