package rdap

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Lookup(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1.2.3.4" {
			t.Errorf("expected path /1.2.3.4, got %s", r.URL.Path)
		}
		resp := rdapResponse{
			Country: "US",
			Name:    "TEST-NET",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL + "/{REMOTE_IP}")
	info, err := client.Lookup(context.Background(), "1.2.3.4")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	if info.Country != "US" {
		t.Errorf("expected country US, got %s", info.Country)
	}
	if info.Name != "TEST-NET" {
		t.Errorf("expected name TEST-NET, got %s", info.Name)
	}
}
