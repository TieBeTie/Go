//go:build !solution

package requestlog

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/felixge/httpsnoop"
	"go.uber.org/zap"
)

var requestCounter int64

func Log(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), atomic.AddInt64(&requestCounter, 1))

			l.Info("request started",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.String("request_id", requestID),
			)

			startTime := time.Now()

			var statusCode int

			logCompletion := func(msg string, extra ...zap.Field) {
				fields := []zap.Field{
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.String("request_id", requestID),
					zap.Duration("duration", time.Since(startTime)),
				}

				if statusCode != 0 {
					fields = append(fields, zap.Int("status_code", statusCode))
				}

				fields = append(fields, extra...)

				l.Info(msg, fields...)
			}

			defer func() {
				if r := recover(); r != nil {
					logCompletion("request panicked")
					panic(r)
				}
			}()

			m := httpsnoop.CaptureMetrics(next, w, r)
			statusCode = m.Code

			logCompletion("request finished")
		})
	}
}
