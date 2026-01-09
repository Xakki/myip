package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		headers  map[string]string
		remote   string
		expected string
	}{
		{
			name:     "RemoteAddr",
			remote:   "1.2.3.4:1234",
			expected: "1.2.3.4",
		},
		{
			name: "X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "10.0.0.1, 1.2.3.4",
			},
			remote:   "192.168.1.1:1234",
			expected: "10.0.0.1",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "10.0.0.2",
			},
			remote:   "192.168.1.1:1234",
			expected: "10.0.0.2",
		},
		{
			name: "X-Forwarded-For takes precedence",
			headers: map[string]string{
				"X-Forwarded-For": "10.0.0.1",
				"X-Real-IP":       "10.0.0.2",
			},
			remote:   "192.168.1.1:1234",
			expected: "10.0.0.1",
		},
		{
			name:     "GET parameter ip",
			url:      "/?ip=8.8.8.8",
			expected: "8.8.8.8",
		},
		{
			name: "GET parameter ip takes precedence",
			url:  "/?ip=8.8.8.8",
			headers: map[string]string{
				"X-Forwarded-For": "10.0.0.1",
			},
			expected: "8.8.8.8",
		},
		{
			name:     "Invalid GET parameter ip fallback",
			url:      "/?ip=invalid",
			remote:   "1.2.3.4:1234",
			expected: "1.2.3.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/"
			if tt.url != "" {
				path = tt.url
			}
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.RemoteAddr = tt.remote
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := clientIP(req)
			if got != tt.expected {
				t.Errorf("clientIP() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestWantsJSON(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		host     string
		headers  map[string]string
		expected bool
	}{
		{
			name:     "Default HTML",
			url:      "/",
			expected: false,
		},
		{
			name: "Content-Type JSON",
			url:  "/",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expected: true,
		},
		{
			name:     "Path /api",
			url:      "/api",
			expected: true,
		},
		{
			name:     "Path /api/something",
			url:      "/api/something",
			expected: true,
		},
		{
			name:     "Subdomain api",
			url:      "/",
			host:     "api.example.com",
			expected: true,
		},
		{
			name:     "Subdomain api with port",
			url:      "/",
			host:     "api.example.com:8080",
			expected: true,
		},
		{
			name:     "Other subdomain",
			url:      "/",
			host:     "www.example.com",
			expected: false,
		},
		{
			name:     "Last level subdomain api",
			url:      "/",
			host:     "my.api",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			if tt.host != "" {
				req.Host = tt.host
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			if got := wantsJSON(req); got != tt.expected {
				t.Errorf("wantsJSON() = %v, want %v", got, tt.expected)
			}
		})
	}
}
