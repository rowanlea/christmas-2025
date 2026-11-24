package viewer

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func doPost(t *testing.T, base, path, body string, hdr bool) *http.Response {
	t.Helper()
	req, _ := http.NewRequest("POST", base+path, strings.NewReader(body))
	if hdr {
		req.Header.Set("X-HTH", "1")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	return resp
}

const pub = `{"ID":"p1","Kind":"run","Cards":[{"ID":"c1","Title":"3.2","Size":16,"FrameCount":2,"IntervalMs":100}]}`

func frameJSON(i int) string {
	b, _ := json.Marshal(FrameEvent{Publication: "p1", Card: "c1", Index: i, Line: "RGB"})
	return string(b)
}

func TestPostsRequireHeader(t *testing.T) {
	srv := httptest.NewServer(Handler())
	defer srv.Close()
	for _, p := range []string{"/publish", "/frame", "/done"} {
		if r := doPost(t, srv.URL, p, "{}", false); r.StatusCode != 403 {
			t.Errorf("%s without header: %d, want 403", p, r.StatusCode)
		}
		if r := doPost(t, srv.URL, p, "{}", true); r.StatusCode != 204 {
			t.Errorf("%s with header: %d, want 204", p, r.StatusCode)
		}
	}
}

func TestLatestIncludesFrames(t *testing.T) {
	srv := httptest.NewServer(Handler())
	defer srv.Close()
	resp, _ := http.Get(srv.URL + "/latest")
	resp.Body.Close()
	if resp.StatusCode != 204 {
		t.Fatalf("empty /latest: %d, want 204", resp.StatusCode)
	}
	doPost(t, srv.URL, "/publish", pub, true)
	doPost(t, srv.URL, "/frame", frameJSON(1), true)
	doPost(t, srv.URL, "/frame", frameJSON(0), true)
	resp, err := http.Get(srv.URL + "/latest")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got struct {
		Publication Publication
		Frames      []FrameEvent
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Publication.ID != "p1" || len(got.Frames) != 2 || got.Frames[0].Index != 0 || got.Frames[1].Index != 1 {
		t.Fatalf("unexpected /latest: %+v", got)
	}
}

type sseClient struct {
	events chan [2]string
	cancel context.CancelFunc
}

func connect(t *testing.T, base string) *sseClient {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "GET", base+"/events", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	c := &sseClient{events: make(chan [2]string, 16), cancel: cancel}
	go func() {
		defer resp.Body.Close()
		sc := bufio.NewScanner(resp.Body)
		var name string
		for sc.Scan() {
			line := sc.Text()
			switch {
			case strings.HasPrefix(line, "event: "):
				name = strings.TrimPrefix(line, "event: ")
			case strings.HasPrefix(line, "data: "):
				c.events <- [2]string{name, strings.TrimPrefix(line, "data: ")}
			}
		}
	}()
	return c
}

func (c *sseClient) expect(t *testing.T, name string) string {
	t.Helper()
	select {
	case e := <-c.events:
		if e[0] != name {
			t.Fatalf("got event %q, want %q", e[0], name)
		}
		return e[1]
	case <-time.After(time.Second):
		t.Fatalf("no %q event within 1s", name)
	}
	return ""
}

func waitClients(t *testing.T, base string, n int) {
	t.Helper()
	for i := 0; i < 100; i++ {
		if resp, err := http.Get(base + "/health"); err == nil {
			var h struct{ Clients int }
			json.NewDecoder(resp.Body).Decode(&h)
			resp.Body.Close()
			if h.Clients == n {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("clients never reached %d", n)
}

func TestEventsStream(t *testing.T) {
	srv := httptest.NewServer(Handler())
	defer srv.Close()
	c := connect(t, srv.URL)
	defer c.cancel()
	waitClients(t, srv.URL, 1)
	doPost(t, srv.URL, "/publish", pub, true)
	if d := c.expect(t, "publication"); !strings.Contains(d, `"p1"`) {
		t.Errorf("publication data: %s", d)
	}
	doPost(t, srv.URL, "/frame", frameJSON(0), true)
	if d := c.expect(t, "frame"); !strings.Contains(d, `"RGB"`) {
		t.Errorf("frame data: %s", d)
	}
	doPost(t, srv.URL, "/done", `{"card":"c1","skippedAtEnd":3}`, true)
	if d := c.expect(t, "done"); !strings.Contains(d, `"c1"`) {
		t.Errorf("done data: %s", d)
	}
}

func TestHealthCountsClients(t *testing.T) {
	srv := httptest.NewServer(Handler())
	defer srv.Close()
	resp, _ := http.Get(srv.URL + "/health")
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if strings.TrimSpace(string(b)) != `{"app":"hth","clients":0}` {
		t.Fatalf("health: %s", b)
	}
	c := connect(t, srv.URL)
	waitClients(t, srv.URL, 1)
	c.cancel()
	waitClients(t, srv.URL, 0)
}

func TestBodyLimit(t *testing.T) {
	srv := httptest.NewServer(Handler())
	defer srv.Close()
	big := `{"ID":"` + strings.Repeat("a", 1<<20) + `"}`
	if r := doPost(t, srv.URL, "/publish", big, true); r.StatusCode != 413 {
		t.Fatalf("oversize body: %d, want 413", r.StatusCode)
	}
}

func TestForeignPortError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, "<html>nope</html>")
	}))
	defer srv.Close()
	t.Setenv("HTH_VIEWER_PORT", "")
	_, err := openAt(strings.TrimPrefix(srv.URL, "http://"), Publication{ID: "p"})
	want := "port 47474 is in use; set HTH_VIEWER_PORT"
	if err == nil || err.Error() != want {
		t.Fatalf("err = %v, want %q", err, want)
	}
}

func TestArtServed(t *testing.T) {
	srv := httptest.NewServer(Handler())
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/art/bulbs-16.json")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("/art/bulbs-16.json: %d", resp.StatusCode)
	}
}
