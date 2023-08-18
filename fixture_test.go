package turtle

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func getTestDuration(t testing.TB) time.Duration {
	t.Helper()

	const envKey = "TURTLE_TEST_DURATION"

	if v := os.Getenv(envKey);v != "" {
		d, err := time.ParseDuration(v)
		require.NoError(t, err, "invalid value for %q", envKey)
		return d
	}

	return 3 * time.Second
}

func getTestConnections(t testing.TB) int {
	t.Helper()

	const envKey = "TURTLE_TEST_CONNECTIONS"

	if v := os.Getenv(envKey);v != "" {
		d, err := strconv.ParseInt(v, 10, 32)
		require.NoError(t, err, "invalid value for %q", envKey)
		return int(d)
	}

	return 10
}

func getTestTarget(t testing.TB, u url.URL) Target {
	t.Helper()

	return Target{
		Url: u,
		Duration: getTestDuration(t),
		Connections: getTestConnections(t),
	}
}

func getTestContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt)
}