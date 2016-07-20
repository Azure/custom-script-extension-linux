package download

import (
	"io"
	"math"
	"time"
)

// SleepFunc pauses the execution for at least duration d.
type SleepFunc func(d time.Duration)

var (
	// ActualSleep uses actual time to pause the execution.
	ActualSleep SleepFunc = time.Sleep
)

const (
	// time to sleep between retries is an exponential backoff formula:
	//   t(n) = k * m^n
	expRetryN = 7 // how many times we retry the Download
	expRetryK = time.Second * 3
	expRetryM = 2
)

// WithRetries retrieves a response body using the specified downloader. Any
// error returned from d will be retried (and retrieved response bodies will be
// closed on failures). If the retries do not succeed, the last error is returned.
//
// It sleeps in exponentially increasing durations between retries.
func WithRetries(d Downloader, sf SleepFunc) (io.ReadCloser, error) {
	var lastErr error
	for n := 0; n < expRetryN; n++ {
		out, err := Download(d)
		if err == nil {
			return out, nil
		}
		lastErr = err

		if out != nil { // we are not going to read this response body
			out.Close()
		}

		if n != expRetryN-1 {
			// have more retries to go, sleep before retrying
			sf(expRetryK * time.Duration(int(math.Pow(float64(expRetryM), float64(n)))))
		}
	}
	return nil, lastErr
}
