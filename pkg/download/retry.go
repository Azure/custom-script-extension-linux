package download

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-extension-platform/vmextension"
	"github.com/go-kit/kit/log"

	errorutil "github.com/Azure/custom-script-extension-linux/pkg/errorutil"
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
	expRetryN    = 7 // how many times we retry the Download
	expRetryK    = time.Second * 3
	expRetryM    = 2
	writeBufSize = 1024 * 8
)

// WithRetries retrieves a response body using the specified downloader. Any
// error returned from d will be retried (and retrieved response bodies will be
// closed on failures). If the retries do not succeed, the last error is returned.
//
// It sleeps in exponentially increasing durations between retries.
func WithRetries(ctx *log.Context, f *os.File, downloaders []Downloader, sf SleepFunc) (int64, *vmextension.ErrorWithClarification) {
	var lastErr error
	var lastErrCode int
	for _, d := range downloaders {
		for n := 0; n < expRetryN; n++ {
			ctx := ctx.With("retry", n)

			// reset the last error before each retry
			lastErr = nil
			lastErrCode = errorutil.NoError
			start := time.Now()
			status, out, ewc := Download(ctx, d)
			if ewc.Err == nil {
				// server returned status code 200 OK
				// we have a response body, copy it to the file
				nBytes, innerErr := io.CopyBuffer(f, out, make([]byte, writeBufSize))
				if innerErr == nil {
					// we are done, close the response body, log time taken to download the file
					// and return the number of bytes written
					out.Close()
					end := time.Since(start)
					ctx.Log("info", fmt.Sprintf("file download sucessful: downloaded and saved %d bytes in %d milliseconds", nBytes, end.Milliseconds()))
					ewc := vmextension.NewErrorWithClarification(lastErrCode, lastErr)
					return nBytes, &ewc
				} else {
					// we failed to download the response body and write it to file
					// because either connection was closed prematurely or file write operation failed
					// mark status as -1 so that we retry
					status = -1
					// clear out the contents of the file so as to not leave a partial file
					f.Truncate(0)
					// cache the inner error
					lastErrCode = errorutil.FileDownload_genericError
					lastErr = innerErr
				}
			} else {
				// cache the outer error
				lastErr = ewc.Err
				lastErrCode = ewc.ErrorCode
			}

			// we are here because either server returned a non-200 status code
			// or we failed to download the response body and write it to file
			// log the error, time elapsed, bytes downloaded, and close the response body
			end := time.Since(start)

			ctx.Log("error", fmt.Sprintf("file download failed with error '%s' : downloaded and saved %d bytes in %d milliseconds", lastErr, 0, end.Milliseconds()))

			if out != nil { // we are not going to read this response body
				out.Close()
			}

			// status == -1 the value when there was no http request
			if status != -1 && !isTransientHttpStatusCode(status) {
				ctx.Log("info", fmt.Sprintf("downloader %T returned %v, skipping retries", d, status))
				break
			}

			if n != expRetryN-1 {
				// have more retries to go, sleep before retrying
				slp := expRetryK * time.Duration(int(math.Pow(float64(expRetryM), float64(n))))
				ctx.Log("sleep", slp)
				sf(slp)
			}
		}
	}

	if lastErr == nil {
		return 0, nil
	}

	ewc := vmextension.NewErrorWithClarification(lastErrCode, lastErr)
	return 0, &ewc
}

func isTransientHttpStatusCode(statusCode int) bool {
	switch statusCode {
	case
		http.StatusRequestTimeout,      // 408
		http.StatusTooManyRequests,     // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true // timeout and too many requests
	default:
		return false
	}
}
