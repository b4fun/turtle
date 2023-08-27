package turtle

import (
	"context"
	"io"
	"net/http"
	"time"
)

type slowReader struct {
	Data <-chan struct{}
}

func startSlowReader(ctx context.Context, timeout time.Duration, randn randn) io.Reader {
	data := make(chan struct{})
	rv := &slowReader{
		Data: data,
	}

	go func() {
		defer close(data)

		var timeoutC <-chan time.Time
		if timeout > 0 {
			timer := time.NewTimer(timeout)
			defer timer.Stop()
			timeoutC = timer.C
		}

		nextSendDataTimer := func() *time.Timer {
			d := time.Duration(randn(100)) * time.Millisecond
			return time.NewTimer(d)
		}

		sendDataTimer := nextSendDataTimer()
		defer func() {
			if sendDataTimer != nil {
				sendDataTimer.Stop()
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-timeoutC:
				return
			case <-sendDataTimer.C:
				data <- struct{}{}
				sendDataTimer = nextSendDataTimer()
			}
		}
	}()

	return rv
}

func (s *slowReader) Read(p []byte) (n int, err error) {
	_, ok := <-s.Data
	if !ok {
		return 0, io.EOF
	}
	if len(p) > 0 {
		p[0] = 'a'
		return 1, nil
	}
	return 0, nil
}

// SlowBodyReadRequest provides the configurations for simulating slow request body reading attack.
type SlowBodyReadRequest struct {
	Target Target `embed:""`

	// Method - the HTTP method to use, one of POST / PUT. Defaults to POST.
	Method string `name:"http-method" help:"the HTTP method to use, one of POST / PUT. Defaults to POST."`

	// BodyReadTimeout - the timeout for reading the request body. Defaults to hang forever.
	BodyReadTimeout time.Duration `name:"http-read-timeout" help:"the timeout for reading the request body. Defaults to hang forever."`

	// randn - the random number generator. For unit testing.
	randn randn
}

func (s *SlowBodyReadRequest) AfterApply() error {
	return s.defaults()
}

func (s *SlowBodyReadRequest) defaults() error {
	if err := s.Target.defaults(); err != nil {
		return err
	}
	if s.Method == "" {
		s.Method = http.MethodPost
	}
	if s.BodyReadTimeout <= 0 {
		s.BodyReadTimeout = 0
	}
	if s.randn == nil {
		s.randn = defaultRandn()
	}

	return nil
}

func (s *SlowBodyReadRequest) Run(ctx context.Context) error {
	if err := s.defaults(); err != nil {
		return err
	}

	workerCtx, cancelWorker := context.WithTimeout(ctx, s.Target.Duration)
	defer cancelWorker()

	runWithWorker(
		workerCtx,
		s.Target.Connections,
		func(ctx context.Context, workerId int) {
			workerEventHandler := wrapEventSettings(s.Target.EventHandler, WithEventWorkerId(workerId))

			defer capturePanic(workerEventHandler)

			err := s.worker(ctx, workerEventHandler)
			if err != nil {
				workerEventHandler.HandleEvent(NewEvent(EventWorkerError, WithEventError(err)))
			}
		},
	)

	return nil
}

func (s *SlowBodyReadRequest) worker(ctx context.Context, eventHandler EventHandler) error {
	body := startSlowReader(ctx, s.BodyReadTimeout, s.randn)

	req, err := http.NewRequestWithContext(ctx, s.Method, s.Target.Url.String(), body)
	if err != nil {
		return err
	}

	client := http.Client{}

	eventHandler.HandleEvent(NewEvent(EventTCPDial))
	defer func() {
		eventHandler.HandleEvent(NewEvent(EventTCPClosed))
	}()

	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}
