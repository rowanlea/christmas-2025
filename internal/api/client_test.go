package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newClient(srv *httptest.Server) Client {
	return Client{
		Config: Config{Server: srv.URL + "/", Team: "team-07", Secret: "k"},
		HTTP:   srv.Client(),
	}
}

func TestRegisterReturnsSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/register/team-07" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Team-Name") != "" || r.Header.Get("X-Unique-Key") != "" {
			t.Error("register must not send auth headers")
		}
		json.NewEncoder(w).Encode("abc123")
	}))
	defer srv.Close()
	secret, err := newClient(srv).Register("team-07")
	if err != nil || secret != "abc123" {
		t.Fatalf("secret=%q err=%v", secret, err)
	}
}

func TestFetchSendsAuthHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/puzzles/3.2/challenges" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Team-Name") != "team-07" || r.Header.Get("X-Unique-Key") != "k" {
			t.Errorf("headers: %v", r.Header)
		}
		w.WriteHeader(201)
		io.WriteString(w, `{"challengeId":"g-1","puzzle":"3.2","size":16,"script":"SIZE 16\n","expiresAt":"2025-12-01T10:00:00Z"}`)
	}))
	defer srv.Close()
	ch, err := newClient(srv).FetchChallenge("3.2")
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID != "g-1" || ch.Puzzle != "3.2" || ch.Size != 16 || ch.Script != "SIZE 16\n" || ch.ExpiresAt.Year() != 2025 {
		t.Fatalf("got %+v", ch)
	}
}

func ptr(s string) *string { return &s }

func TestSubmitPostsPNG(t *testing.T) {
	for name, tc := range map[string]struct {
		body string
		want *string
	}{
		"string": {`"3.3"`, ptr("3.3")},
		"null":   {`null`, nil},
	} {
		t.Run(name, func(t *testing.T) {
			png := []byte{0x89, 'P', 'N', 'G', 1, 2, 3}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.URL.Path != "/challenges/g-1/submission" {
					t.Errorf("got %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("Content-Type") != "image/png" {
					t.Errorf("content-type %q", r.Header.Get("Content-Type"))
				}
				if r.Header.Get("X-Team-Name") != "team-07" || r.Header.Get("X-Unique-Key") != "k" {
					t.Errorf("headers: %v", r.Header)
				}
				b, _ := io.ReadAll(r.Body)
				if string(b) != string(png) {
					t.Errorf("body %v", b)
				}
				io.WriteString(w, `{"puzzle":"3.2","correct":131,"bulbCount":136,"points":72,"bestPoints":72,"unlocked":`+tc.body+`}`)
			}))
			defer srv.Close()
			s, err := newClient(srv).Submit("g-1", png)
			if err != nil {
				t.Fatal(err)
			}
			if s.Puzzle != "3.2" || s.Correct != 131 || s.BulbCount != 136 || s.Points != 72 || s.BestPoints != 72 {
				t.Fatalf("got %+v", s)
			}
			if (s.Unlocked == nil) != (tc.want == nil) || (tc.want != nil && *s.Unlocked != *tc.want) {
				t.Fatalf("unlocked = %v", s.Unlocked)
			}
		})
	}
}

func TestShowPermission(t *testing.T) {
	const msg = "Santa comes first: get full marks on 7.1 to unlock light shows"
	status := 204
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/shows/permission" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(status)
		if status == 403 {
			json.NewEncoder(w).Encode(map[string]string{"error": msg})
		}
	}))
	defer srv.Close()
	c := newClient(srv)
	if err := c.ShowPermission(); err != nil {
		t.Fatal(err)
	}
	status = 403
	err := c.ShowPermission()
	var ae *APIError
	if !errors.As(err, &ae) || ae.Status != 403 || err.Error() != msg {
		t.Fatalf("err = %v", err)
	}
}

func TestUploadShowPostsFrames(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/shows" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" || r.Header.Get("X-Team-Name") != "team-07" {
			t.Errorf("headers: %v", r.Header)
		}
		var body struct{ Frames []string }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Frames) != 2 || body.Frames[1] != "B" {
			t.Errorf("body %+v err %v", body, err)
		}
		w.WriteHeader(201)
		io.WriteString(w, `{"frames":2}`)
	}))
	defer srv.Close()
	if err := newClient(srv).UploadShow([]string{"R", "B"}); err != nil {
		t.Fatal(err)
	}
}

func TestErrorMapping(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   string
	}{
		{401, `{"error":"bad"}`, "not logged in or wrong secret; run hth login <team> --secret <secret>"},
		{403, `{"error":"puzzle locked"}`, "puzzle locked"},
		{404, `{"error":"unknown puzzle"}`, "unknown puzzle"},
		{409, `{"error":"x"}`, "challenge already submitted; run again for a fresh one"},
		{410, `{"error":"x"}`, "challenge expired (5 minutes); run again for a fresh one"},
		{400, `{"error":"a show has 1 to 500 frames"}`, "request refused (400): a show has 1 to 500 frames"},
		{413, `{"error":"too big"}`, "request refused (413): too big"},
		{429, `{"error":"slow down"}`, "request refused (429): slow down"},
		{500, `{"error":"boom"}`, "server error 500: boom"},
		{502, "  bad gateway \n", "server error 502: bad gateway"},
	}
	for _, tc := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
			io.WriteString(w, tc.body)
		}))
		_, err := newClient(srv).FetchChallenge("1.1")
		srv.Close()
		var ae *APIError
		if !errors.As(err, &ae) || ae.Status != tc.status {
			t.Fatalf("%d: err = %v", tc.status, err)
		}
		if err.Error() != tc.want {
			t.Errorf("%d: got %q, want %q", tc.status, err.Error(), tc.want)
		}
	}
}

func TestNetworkErrorWrapped(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	c := newClient(srv)
	srv.Close()
	_, err := c.FetchChallenge("1.1")
	if err == nil || !strings.HasPrefix(err.Error(), "could not reach "+c.Server+": ") {
		t.Fatalf("err = %v", err)
	}
}
