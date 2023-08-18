package turtle

import (
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/b4fun/turtle/internal/testhttp"
	"github.com/stretchr/testify/assert"
)

func Test_SlowBodyReadRequest_defaults(t *testing.T) {
	v := SlowBodyReadRequest{}

	assert.NoError(t, v.defaults())

	assert.Equal(t, http.MethodPost, v.Method)
	assert.NotNil(t, v.randn)
}

func Test_SlowBodyReadRequest_vulnerable(t *testing.T) {
	t.Parallel()

	counter := testhttp.NewConnStateCounters()

	testCtx, stopTest := getTestContext()
	defer stopTest()

	serverUrl, serverStopped := testhttp.CreateTestServer(testCtx, t, func(s *http.Server) {
		s.ReadTimeout = 0 // hung forever
		s.ConnState = counter.ServerConnState
		s.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil {
				return
			}
			defer r.Body.Close()

			_, err := io.Copy(io.Discard, r.Body)
			if err != nil {
				assert.True(t, os.IsTimeout(err), "read body should return timeout error")
			}
		})
	})

	target := getTestTarget(t, *serverUrl)
	attack := SlowBodyReadRequest{Target: target}

	assert.NoError(t, attack.Run(testCtx))
	stopTest()
	<-serverStopped

	t.Log(counter.String())

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

func Test_SlowBodyReadRequest_invulnerable(t *testing.T) {
	t.Parallel()

	counter := testhttp.NewConnStateCounters()
	timeoutErrSeen := new(atomic.Bool)

	testCtx, stopTest := getTestContext()
	defer stopTest()

	serverUrl, serverStopped := testhttp.CreateTestServer(testCtx, t, func(s *http.Server) {
		s.ReadTimeout = 1 * time.Second
		s.ConnState = counter.ServerConnState
		s.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil {
				return
			}
			defer r.Body.Close()

			_, err := io.Copy(io.Discard, r.Body)
			if err != nil {
				assert.True(t, os.IsTimeout(err), "read body should return timeout error")
				timeoutErrSeen.Store(true)
			}
		})
	})

	target := getTestTarget(t, *serverUrl)
	attack := SlowBodyReadRequest{Target: target}

	assert.NoError(t, attack.Run(testCtx))
	stopTest()
	<-serverStopped

	t.Log(counter.String())

	conns := counter.GetConns()
	assert.NotEmpty(t, conns, "should have at least one connection")
	assert.Greater(t, len(conns), target.Connections, "should create more connections")
	for _, conn := range conns {
		timeline := counter.GetConnStateTimeline(conn)
		assert.GreaterOrEqual(t, len(timeline), 1, "should have at least one state")
		// a typical state transition is: new - (read body) -> active - (read body timeout close) -> closed
		assert.Equal(t, http.StateNew, timeline[0])
	}
	assert.True(t, timeoutErrSeen.Load(), "should see at least one timeout error")
}