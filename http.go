package turtle

import (
	crand "crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"time"
)

type randn func(int) int

func defaultRandn() randn {
	return func(n int) int {
		v, err := crand.Int(crand.Reader, big.NewInt(int64(n)))
		if err != nil {
			panic(fmt.Sprintf("failed to generate random number: %v", err))
		}
		return int(v.Int64())
	}
}

type tcpConnWithWriteTimeout struct {
	conn         net.Conn
	writeTimeout time.Duration
}

var _ io.Writer = (*tcpConnWithWriteTimeout)(nil)

func (w *tcpConnWithWriteTimeout) Write(b []byte) (int, error) {
	if err := w.conn.SetWriteDeadline(time.Now().Add(w.writeTimeout)); err != nil {
		return 0, fmt.Errorf("SetWriteDeadline: %w", err)
	}

	return w.conn.Write(b)
}

var defaultUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.38 (KHTML, like Gecko) Chrome/47.0.3104.383 Safari/603",
	"Mozilla/5.0 (Linux; U; Android 4.4; LG-V710 Build/KOT49I) AppleWebKit/601.6 (KHTML, like Gecko)  Chrome/48.0.1587.379 Mobile Safari/537.5",
	"Mozilla/5.0 (Windows NT 10.4; Win64; x64) AppleWebKit/535.22 (KHTML, like Gecko) Chrome/50.0.2318.242 Safari/536",
	"Mozilla/5.0 (Windows; Windows NT 10.1; Win64; x64; en-US) Gecko/20130401 Firefox/66.1",
}

func httpLine(s string) string {
	return s + "\r\n"
}

func writeHTTPLineTo(w io.Writer, s string) error {

	l := httpLine(s)
	fmt.Println("writing ", l)
	n, err := w.Write([]byte(l))
	if err != nil {
		return err
	}
	if len(l) != n {
		return io.ErrShortWrite
	}

	return nil
}

func writeHTTPHeaderTo(w io.Writer, key, value string) error {
	return writeHTTPLineTo(w, fmt.Sprintf("%s: %s", http.CanonicalHeaderKey(key), value))
}

func gibberishValue(randn randn, size int) string {
	const firstLetters = "abcdefghiklmnopqrstuvwxyz0123456789"
	const letters = firstLetters + "-"

	b := make([]byte, size)
	b[0] = firstLetters[randn(len(firstLetters))]
	b[size-1] = firstLetters[randn(len(firstLetters))]
	for i := 1; i <= size-1; i++ {
		b[i] = letters[randn(len(letters))]
	}

	return string(b)
}
