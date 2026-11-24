package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	Config
	HTTP *http.Client
}

// NewClient returns a client with a 30 s HTTP timeout.
func NewClient(cfg Config) Client {
	return Client{Config: cfg, HTTP: &http.Client{Timeout: 30 * time.Second}}
}

type Challenge struct {
	ID        string    `json:"challengeId"`
	Puzzle    string    `json:"puzzle"`
	Size      int       `json:"size"`
	Script    string    `json:"script"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type Score struct {
	Puzzle     string  `json:"puzzle"`
	Correct    int     `json:"correct"`
	BulbCount  int     `json:"bulbCount"`
	Points     int     `json:"points"`
	BestPoints int     `json:"bestPoints"`
	Unlocked   *string `json:"unlocked"`
}

type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	switch e.Status {
	case 401:
		return "not logged in or wrong secret; run hth login <team> --secret <secret>"
	case 403, 404:
		return e.Message
	case 409:
		return "challenge already submitted; run again for a fresh one"
	case 410:
		return "challenge expired (5 minutes); run again for a fresh one"
	}
	if e.Status >= 400 && e.Status < 500 {
		return fmt.Sprintf("request refused (%d): %s", e.Status, e.Message)
	}
	return fmt.Sprintf("server error %d: %s", e.Status, e.Message)
}

func (c Client) Register(team string) (string, error) {
	var secret string
	err := c.do("GET", "/register/"+url.PathEscape(team), "", nil, false, 200, &secret)
	return secret, err
}

func (c Client) FetchChallenge(puzzle string) (Challenge, error) {
	var ch Challenge
	err := c.do("POST", "/puzzles/"+url.PathEscape(puzzle)+"/challenges", "", nil, true, 201, &ch)
	return ch, err
}

func (c Client) Submit(id string, png []byte) (Score, error) {
	var s Score
	err := c.do("POST", "/challenges/"+url.PathEscape(id)+"/submission", "image/png", png, true, 200, &s)
	return s, err
}

func (c Client) ShowPermission() error {
	return c.do("GET", "/shows/permission", "", nil, true, 204, nil)
}

func (c Client) UploadShow(frames []string) error {
	body, err := json.Marshal(map[string][]string{"frames": frames})
	if err != nil {
		return err
	}
	return c.do("POST", "/shows", "application/json", body, true, 201, nil)
}

func (c Client) do(method, path, contentType string, body []byte, auth bool, okStatus int, out any) error {
	server := strings.TrimRight(c.Server, "/")
	req, err := http.NewRequest(method, server+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if auth {
		req.Header.Set("X-Team-Name", c.Team)
		req.Header.Set("X-Unique-Key", c.Secret)
	}
	httpc := c.HTTP
	if httpc == nil {
		httpc = NewClient(c.Config).HTTP
	}
	resp, err := httpc.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach %s: %w", c.Server, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not reach %s: %w", c.Server, err)
	}
	if resp.StatusCode != okStatus {
		return &APIError{Status: resp.StatusCode, Message: errorMessage(data)}
	}
	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("unexpected response from %s: %w", c.Server, err)
		}
	}
	return nil
}

func errorMessage(body []byte) string {
	var e struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &e) == nil && e.Error != "" {
		return e.Error
	}
	return strings.TrimSpace(string(body))
}
