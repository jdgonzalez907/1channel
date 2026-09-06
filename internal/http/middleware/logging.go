package middleware

import (
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	cfConnectingIPHeader = "CF-Connecting-IP"
	userAgentHeader      = "User-Agent"
)

func Logging() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UTC()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			bodyCounter := &countingBody{ReadCloser: r.Body}
			r.Body = bodyCounter

			next.ServeHTTP(recorder, r)

			slog.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.status,
				"request_bytes", requestSize(r, bodyCounter),
				"response_bytes", recorder.size,
				"client_ip", clientIP(r),
				"user_agent", r.UserAgent(),
				"duration", time.Since(start),
			)
		})
	}
}

// requestSize returns the declared size when the client sent Content-Length,
// otherwise the number of body bytes actually consumed by the handler.
func requestSize(r *http.Request, counter *countingBody) int64 {
	if r.ContentLength >= 0 {
		return r.ContentLength
	}
	return counter.read.Load()
}

// clientIP prefers the real client address set by the Cloudflare edge
// (RemoteAddr behind the tunnel is always loopback/bridge), falling back to
// RemoteAddr without its port.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get(cfConnectingIPHeader); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type countingBody struct {
	io.ReadCloser
	read atomic.Int64
}

func (c *countingBody) Read(bytes []byte) (int, error) {
	n, err := c.ReadCloser.Read(bytes)
	c.read.Add(int64(n))
	return n, err
}

// statusRecorder wraps the real ResponseWriter to capture the final status and
// response body size. Known limitation: it hides optional interfaces like
// http.Flusher/http.Hijacker; acceptable here because http.TimeoutHandler in
// the chain already buffers responses and blocks hijacking.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	size        int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if !r.wroteHeader {
		r.status = status
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(bytes []byte) (int, error) {
	if !r.wroteHeader {
		r.wroteHeader = true
	}
	n, err := r.ResponseWriter.Write(bytes)
	r.size += n
	return n, err
}
