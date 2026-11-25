// Package viewer is the local daemon that streams tree frames to a browser.
package viewer

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"sort"
	"sync"

	"hth/internal/art"
)

//go:embed static/*
var static embed.FS

const maxBody = 1 << 20

// Card is one tree, with any expected tree beside it.
type Card struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Size       int    `json:"size"`
	Answer     string `json:"answer"`
	Expected   string `json:"expected"` // test only
	Summary    string `json:"summary"`
	FrameCount int    `json:"frameCount"`
	IntervalMs int    `json:"intervalMs"`
}

// Publication is one set of cards shown together.
type Publication struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"` // "run", "test" or "show"
	Cards []Card `json:"cards"`
}

// FrameEvent is one sampled tree of a card.
type FrameEvent struct {
	Publication   string `json:"publication"`
	Card          string `json:"card"`
	Index         int    `json:"index"`
	Line          string `json:"line"`
	SkippedBefore int64  `json:"skippedBefore"`
}

// DoneEvent marks the end of a card's frames.
type DoneEvent struct {
	Publication  string `json:"publication"`
	Card         string `json:"card"`
	SkippedAtEnd int64  `json:"skippedAtEnd"`
}

// Addr is the daemon's listen address.
func Addr() string {
	if p := os.Getenv("HTH_VIEWER_PORT"); p != "" {
		return "127.0.0.1:" + p
	}
	return "127.0.0.1:47474"
}

// Serve runs the daemon on addr until it fails.
func Serve(addr string) error {
	return (&http.Server{Addr: addr, Handler: Handler()}).ListenAndServe()
}

type event struct{ name, data string }

type state struct {
	mu      sync.Mutex
	pub     *Publication
	frames  []FrameEvent
	done    []DoneEvent
	clients map[chan event]struct{}
}

// Handler returns a fresh daemon handler with its own state.
func Handler() http.Handler {
	s := &state{clients: map[chan event]struct{}{}}
	mux := http.NewServeMux()
	staticFS, _ := fs.Sub(static, "static")
	files := http.FileServerFS(staticFS)
	mux.Handle("GET /{$}", files)
	mux.Handle("GET /static/", http.StripPrefix("/static/", files))
	mux.Handle("GET /art/", http.StripPrefix("/art/", http.FileServerFS(art.Assets())))
	mux.HandleFunc("GET /latest", s.latest)
	mux.HandleFunc("GET /events", s.events)
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("POST /publish", s.publish)
	mux.HandleFunc("POST /frame", s.frame)
	mux.HandleFunc("POST /done", s.doneH)
	return localOnly(mux)
}

// localOnly rejects requests addressed to any host name but a loopback one. The daemon
// listens on 127.0.0.1 only, but a web page can rebind its own domain to 127.0.0.1 and
// then talk to the daemon as same-origin; such requests still carry that domain as Host.
func localOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		if host != "127.0.0.1" && host != "localhost" && host != "::1" && host != "[::1]" {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *state) health(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	n := len(s.clients)
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"app":"hth","clients":%d}`+"\n", n)
}

func (s *state) latest(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if s.pub == nil {
		s.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
		return
	}
	out := struct {
		Publication Publication  `json:"publication"`
		Frames      []FrameEvent `json:"frames"`
		Done        []DoneEvent  `json:"done"`
	}{*s.pub, append([]FrameEvent{}, s.frames...), append([]DoneEvent{}, s.done...)}
	s.mu.Unlock()
	order := map[string]int{}
	for i, c := range out.Publication.Cards {
		order[c.ID] = i
	}
	sort.SliceStable(out.Frames, func(i, j int) bool {
		a, b := out.Frames[i], out.Frames[j]
		if order[a.Card] != order[b.Card] {
			return order[a.Card] < order[b.Card]
		}
		return a.Index < b.Index
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (s *state) events(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	ch := make(chan event, 1024)
	s.mu.Lock()
	s.clients[ch] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		s.mu.Unlock()
	}()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	fl.Flush()
	for {
		select {
		case <-r.Context().Done():
			return
		case e := <-ch:
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.name, e.data)
			fl.Flush()
		}
	}
}

// broadcast must be called with s.mu held.
func (s *state) broadcast(name string, v any) {
	b, _ := json.Marshal(v)
	for ch := range s.clients {
		select {
		case ch <- event{name, string(b)}:
		default: // a stalled browser misses events rather than blocking the CLI
		}
	}
}

// decode checks the guard header and body size, then decodes the JSON body.
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if r.Header.Get("X-HTH") != "1" {
		http.Error(w, "missing X-HTH header", http.StatusForbidden)
		return false
	}
	err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody)).Decode(v)
	var tooBig *http.MaxBytesError
	switch {
	case errors.As(err, &tooBig):
		http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
		return false
	case err != nil:
		http.Error(w, "bad JSON", http.StatusBadRequest)
		return false
	}
	return true
}

func (s *state) publish(w http.ResponseWriter, r *http.Request) {
	var p Publication
	if !decode(w, r, &p) {
		return
	}
	s.mu.Lock()
	s.pub, s.frames, s.done = &p, nil, nil
	s.broadcast("publication", p)
	s.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (s *state) frame(w http.ResponseWriter, r *http.Request) {
	var f FrameEvent
	if !decode(w, r, &f) {
		return
	}
	s.mu.Lock()
	if s.pub != nil && s.pub.ID == f.Publication {
		s.frames = append(s.frames, f)
		s.broadcast("frame", f)
	}
	s.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (s *state) doneH(w http.ResponseWriter, r *http.Request) {
	var d DoneEvent
	if !decode(w, r, &d) {
		return
	}
	s.mu.Lock()
	if s.pub != nil && s.pub.ID == d.Publication {
		s.done = append(s.done, d)
		s.broadcast("done", d)
	}
	s.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}
