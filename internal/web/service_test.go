package web

import (
	"context"
	"testing"
	"time"

	"myip/internal/rdap"
)

type mockRDAPLookup struct {
	lookupFunc func(ctx context.Context, ip string) (rdap.Info, error)
}

func (m *mockRDAPLookup) Lookup(ctx context.Context, ip string) (rdap.Info, error) {
	return m.lookupFunc(ctx, ip)
}

type mockStore struct {
	getCached      func(ctx context.Context, ip string) (rdap.Info, time.Time, bool, error)
	setCached      func(ctx context.Context, ip string, info rdap.Info, fetchedAt time.Time) error
	incrementCount func(ctx context.Context, ip string) (int64, error)
}

func (m *mockStore) GetCached(ctx context.Context, ip string) (rdap.Info, time.Time, bool, error) {
	return m.getCached(ctx, ip)
}
func (m *mockStore) SetCached(ctx context.Context, ip string, info rdap.Info, fetchedAt time.Time) error {
	return m.setCached(ctx, ip, info, fetchedAt)
}
func (m *mockStore) IncrementCount(ctx context.Context, ip string) (int64, error) {
	return m.incrementCount(ctx, ip)
}

func (s *mockService) Fetch(ctx context.Context, ip string) (Response, error) {
	return s.fetch(ctx, ip)
}

func (s *mockService) OnError(err error) {
	s.onError(err)
}

type mockService struct {
	fetch   func(ctx context.Context, ip string) (Response, error)
	onError func(err error)
}

func TestServiceImpl_Fetch(t *testing.T) {
	ms := &mockStore{
		incrementCount: func(ctx context.Context, ip string) (int64, error) {
			return 1, nil
		},
		getCached: func(ctx context.Context, ip string) (rdap.Info, time.Time, bool, error) {
			return rdap.Info{Country: "US"}, time.Now(), true, nil
		},
	}
	ml := &mockRDAPLookup{}

	s := NewService(ms, ml, nil)
	resp, err := s.Fetch(context.Background(), "1.2.3.4")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if resp.IP != "1.2.3.4" {
		t.Errorf("expected IP 1.2.3.4, got %s", resp.IP)
	}
	if resp.CountCall != 1 {
		t.Errorf("expected count 1, got %d", resp.CountCall)
	}
	if resp.RDAP.Country != "US" {
		t.Errorf("expected country US, got %s", resp.RDAP.Country)
	}
}
