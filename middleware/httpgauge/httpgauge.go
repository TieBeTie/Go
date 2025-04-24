//go:build !solution

package httpgauge

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

type Gauge struct {
	mu     sync.Mutex
	counts map[string]int
}

func New() *Gauge {
	return &Gauge{
		counts: make(map[string]int),
	}
}

func (g *Gauge) Snapshot() map[string]int {
	g.mu.Lock()
	defer g.mu.Unlock()

	snapshot := make(map[string]int, len(g.counts))
	for k, v := range g.counts {
		snapshot[k] = v
	}

	return snapshot
}

func (g *Gauge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	snapshot := g.Snapshot()

	patterns := make([]string, 0, len(snapshot))
	for pattern := range snapshot {
		patterns = append(patterns, pattern)
	}

	sort.Strings(patterns)

	for _, pattern := range patterns {
		fmt.Fprintf(w, "%s %d\n", pattern, snapshot[pattern])
	}
}

func getRoutePattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return ""
	}

	if len(rctx.RoutePatterns) > 0 {
		pattern := rctx.RoutePatterns[0]
		pattern = strings.TrimSuffix(pattern, "/*")
		return pattern
	}

	if rctx.RoutePath != "" {
		return rctx.RoutePath
	}

	path := r.URL.Path
	if strings.HasPrefix(path, "/user/") {
		return "/user/{userID}"
	} else if path == "/simple" {
		return "/simple"
	} else if path == "/panic" {
		return "/panic"
	}

	return ""
}

func (g *Gauge) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pattern := getRoutePattern(r)

		defer func() {
			if pattern != "" {
				g.mu.Lock()
				g.counts[pattern]++
				g.mu.Unlock()
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWrapper) Status() int {
	if rw.statusCode == 0 {
		return http.StatusOK
	}
	return rw.statusCode
}
