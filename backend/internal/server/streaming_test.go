package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Regression: LoggingMiddleware and AuditMiddleware wrap the ResponseWriter to
// capture the status code. Embedding http.ResponseWriter promotes only
// Header/Write/WriteHeader, so a wrapper that adds nothing else hides the
// underlying http.Flusher — which made every SSE request to /api/events fail
// the `w.(http.Flusher)` check in EventsHandler.Stream and return
// "streaming not supported". The missing Unwrap additionally turned
// http.ResponseController.SetWriteDeadline into a silent no-op, leaving streams
// subject to the server's WriteTimeout.
//
// Runs against a real connection because httptest.ResponseRecorder does not
// itself support write deadlines.
func TestResponseWrappersStayTransparentToStreaming(t *testing.T) {
	tests := []struct {
		name string
		wrap func(http.Handler) http.Handler
	}{
		{"LoggingMiddleware", LoggingMiddleware},
		{"AuditMiddleware", func(next http.Handler) http.Handler {
			return AuditMiddleware(nil, testResolver(t))(next)
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// The handler runs on the server's goroutine, so hand the results
			// back over a channel: the receive gives us a happens-before edge
			// the race detector can see. Reading shared variables directly
			// races with the handler still finishing after the client has its
			// response.
			type probe struct {
				gotFlusher bool
				deadlineOK bool
			}
			results := make(chan probe, 1)

			inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Non-2xx so AuditMiddleware skips its (here unconfigured) log
				// write; this test is about writer transparency, not auditing.
				w.WriteHeader(http.StatusInternalServerError)

				flusher, ok := w.(http.Flusher)
				if ok {
					flusher.Flush()
				}

				rc := http.NewResponseController(w)
				results <- probe{
					gotFlusher: ok,
					deadlineOK: rc.SetWriteDeadline(time.Now().Add(60*time.Second)) == nil,
				}
			})

			srv := httptest.NewServer(tc.wrap(inner))
			defer srv.Close()

			resp, err := srv.Client().Post(srv.URL+"/api/events", "application/json", nil)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()

			got := <-results

			if !got.gotFlusher {
				t.Error("handler could not assert http.Flusher: SSE would 500")
			}
			if !got.deadlineOK {
				t.Error("ResponseController.SetWriteDeadline did not reach the underlying writer")
			}
		})
	}
}

// The wrappers must expose the writer they wrap so http.ResponseController can
// unwrap through them.
func TestResponseWrappersImplementUnwrap(t *testing.T) {
	base := httptest.NewRecorder()

	wrappers := map[string]http.ResponseWriter{
		"statusWriter":   &statusWriter{ResponseWriter: base},
		"statusRecorder": &statusRecorder{ResponseWriter: base},
	}

	for name, w := range wrappers {
		t.Run(name, func(t *testing.T) {
			unwrapper, ok := w.(interface{ Unwrap() http.ResponseWriter })
			if !ok {
				t.Fatal("wrapper does not implement Unwrap() http.ResponseWriter")
			}
			if unwrapper.Unwrap() != http.ResponseWriter(base) {
				t.Error("Unwrap did not return the wrapped ResponseWriter")
			}
		})
	}
}
