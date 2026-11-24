package viewer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
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
	queue  chan post
	failed bool
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
		return nil, fmt.Errorf("port %s is in use; set HTH_VIEWER_PORT", portOf(Addr()))
	}
	if err := send(base, "/publish", p); err != nil {
		return nil, err
	}
	if clients == 0 {
		if err := openBrowser(base); err != nil {
			return nil, fmt.Errorf("open %s in your browser", base)
		}
	}
	s := &Session{base: base, queue: make(chan post, 1024)}
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
		if !s.failed && send(s.base, p.path, p.body) != nil {
			s.failed = true
		}
		s.wg.Done()
	}
}

func (s *Session) enqueue(path string, v any) {
	s.wg.Add(1)
	s.queue <- post{path, v}
}

// Frame posts one frame in the background.
func (s *Session) Frame(f FrameEvent) { s.enqueue("/frame", f) }

// Done marks a card's frames complete.
func (s *Session) Done(card string, skippedAtEnd int64) {
	s.enqueue("/done", DoneEvent{Card: card, SkippedAtEnd: skippedAtEnd})
}

// Flush waits for queued posts, so a short-lived CLI does not exit before they
// are sent.
func (s *Session) Flush() { s.wg.Wait() }

func portOf(addr string) string { return addr[strings.LastIndex(addr, ":")+1:] }
