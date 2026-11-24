package viewer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	spawnDaemon = spawn
	openBrowser = browse
	httpClient  = &http.Client{Timeout: 2 * time.Second}
)

// Session streams one publication's frames to the daemon. Posts happen in the
// background; after the first failure the rest are dropped.
type Session struct {
	base   string
	pubID  string
	queue  chan post
	failed atomic.Bool
	wg     sync.WaitGroup
}

type post struct {
	path string
	body any
}

// Open starts the daemon if needed, publishes p, and opens the browser when no
// tab is connected.
func Open(p Publication) (*Session, error) { return openAt(Addr(), p) }

func openAt(addr string, p Publication) (*Session, error) {
	base := "http://" + addr
	clients, state := health(base)
	if state == unreachable {
		if err := spawnDaemon(); err != nil {
			return nil, fmt.Errorf("start viewer: %w", err)
		}
		for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); time.Sleep(50 * time.Millisecond) {
			if clients, state = health(base); state != unreachable {
				break
			}
		}
	}
	switch state {
	case unreachable:
		return nil, errors.New("viewer did not start; run hth viewer in another terminal")
	case foreign:
		return nil, fmt.Errorf("port %s is in use; set HTH_VIEWER_PORT", portOf(addr))
	}
	if err := send(base, "/publish", p); err != nil {
		return nil, err
	}
	if clients == 0 {
		if err := openBrowser(base); err != nil {
			return nil, fmt.Errorf("open %s in your browser", base)
		}
	}
	s := &Session{base: base, pubID: p.ID, queue: make(chan post, 1024)}
	go s.run()
	return s, nil
}

type healthState int

const (
	unreachable healthState = iota
	foreign
	ok
)

func health(base string) (int, healthState) {
	resp, err := httpClient.Get(base + "/health")
	if err != nil {
		return 0, unreachable
	}
	defer resp.Body.Close()
	var h struct {
		App     string `json:"app"`
		Clients int    `json:"clients"`
	}
	if resp.StatusCode != 200 || json.NewDecoder(resp.Body).Decode(&h) != nil || h.App != "hth" {
		return 0, foreign
	}
	return h.Clients, ok
}

func send(base, path string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("POST", base+path, bytes.NewReader(b))
	req.Header.Set("X-HTH", "1")
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("viewer: %s", resp.Status)
	}
	return nil
}

func (s *Session) run() {
	for p := range s.queue {
		if !s.failed.Load() && send(s.base, p.path, p.body) != nil {
			s.failed.Store(true)
		}
		s.wg.Done()
	}
}

// enqueue never blocks: when the queue is full the post is dropped and the
// session is marked failed.
func (s *Session) enqueue(path string, v any) {
	if s.failed.Load() {
		return
	}
	s.wg.Add(1)
	select {
	case s.queue <- post{path, v}:
	default:
		s.wg.Done()
		s.failed.Store(true)
	}
}

// Frame posts one frame in the background. It must not be called after Close.
func (s *Session) Frame(f FrameEvent) {
	f.Publication = s.pubID
	s.enqueue("/frame", f)
}

// Done marks a card's frames complete. It must not be called after Close.
func (s *Session) Done(card string, skippedAtEnd int64) {
	s.enqueue("/done", DoneEvent{Publication: s.pubID, Card: card, SkippedAtEnd: skippedAtEnd})
}

// Flush waits for queued posts, so a short-lived CLI does not exit before they
// are sent. Call it only after all Frame and Done calls have returned.
func (s *Session) Flush() { s.wg.Wait() }

// Close flushes, then stops the background goroutine. Call it only after all
// Frame and Done calls have returned, and do not call Frame or Done afterwards.
func (s *Session) Close() {
	s.Flush()
	close(s.queue)
}

func portOf(addr string) string { return addr[strings.LastIndex(addr, ":")+1:] }
