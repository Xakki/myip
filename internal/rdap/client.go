package rdap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const requestTimeout = 5 * time.Second

// Info represents selected RDAP fields.
type Info struct {
	Country   string  `json:"country"`
	Handle    string  `json:"handle"`
	IPVersion string  `json:"ipVersion"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Events    []Event `json:"events"`
}

// Event represents a single RDAP event entry.
type Event struct {
	Action string `json:"action"`
	Date   string `json:"date"`
}

type rdapResponse struct {
	Country   string `json:"country"`
	Handle    string `json:"handle"`
	IPVersion string `json:"ipVersion"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Events    []struct {
		Action string `json:"eventAction"`
		Date   string `json:"eventDate"`
	} `json:"events"`
}

// Client fetches RDAP information for IP addresses.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient creates a new RDAP client using the provided URL template.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// Lookup fetches RDAP information for the given IP.
func (c *Client) Lookup(ctx context.Context, ip string) (Info, error) {
	url := strings.ReplaceAll(c.baseURL, "{REMOTE_IP}", ip)
	if url == c.baseURL {
		return Info{}, fmt.Errorf("RDAP API template missing {REMOTE_IP}")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Info{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return Info{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Info{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var payload rdapResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Info{}, fmt.Errorf("decode response: %w", err)
	}

	info := Info{
		Country:   payload.Country,
		Handle:    payload.Handle,
		IPVersion: payload.IPVersion,
		Name:      payload.Name,
		Type:      payload.Type,
		Events:    make([]Event, 0, len(payload.Events)),
	}

	for _, event := range payload.Events {
		info.Events = append(info.Events, Event{
			Action: event.Action,
			Date:   event.Date,
		})
	}

	return info, nil
}
