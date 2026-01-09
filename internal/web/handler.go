package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
	"time"

	"myip/internal/rdap"
	"myip/internal/store"
)

const requestTimeout = 3 * time.Second

// Service defines the dependencies needed by the HTTP handler.
type Service interface {
	Fetch(ctx context.Context, ip string) (Response, error)
	OnError(err error)
}

// Response represents the data returned for a client request.
type Response struct {
	IP        string       `json:"ip"`
	CountCall int64        `json:"count_call"`
	RDAP      rdap.Info    `json:"rdap"`
	Events    []rdap.Event `json:"events"`
	Error     error        `json:"-"`
}

// Handler serves the root endpoint.
type Handler struct {
	tmpl    *template.Template
	service Service
}

// NewHandler constructs a new Handler.
func NewHandler(tmpl *template.Template, service Service) *Handler {
	return &Handler{tmpl: tmpl, service: service}
}

// ServeHTTP handles the root endpoint.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/api") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ip := clientIP(r)
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	response, err := h.service.Fetch(ctx, ip)
	if err != nil {
		h.service.OnError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if response.Error != nil {
		h.service.OnError(response.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		payload := apiResponse{
			IP:        response.IP,
			CountCall: response.CountCall,
			Country:   response.RDAP.Country,
			Handle:    response.RDAP.Handle,
			IPVersion: response.RDAP.IPVersion,
			Name:      response.RDAP.Name,
			Type:      response.RDAP.Type,
			Events:    response.RDAP.Events,
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			h.service.OnError(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := templateData{
		IP:        response.IP,
		CountCall: response.CountCall,
		RDAP:      response.RDAP,
		HasRDAP:   hasRDAP(response.RDAP),
	}
	if err := h.tmpl.Execute(w, data); err != nil {
		h.service.OnError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type apiResponse struct {
	IP        string       `json:"ip"`
	CountCall int64        `json:"count_call"`
	Country   string       `json:"country"`
	Handle    string       `json:"handle"`
	IPVersion string       `json:"ipVersion"`
	Name      string       `json:"name"`
	Type      string       `json:"type"`
	Events    []rdap.Event `json:"events"`
}

type templateData struct {
	IP        string
	CountCall int64
	RDAP      rdap.Info
	HasRDAP   bool
}

func clientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.URL.Query().Get("ip")); ip != "" {
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		for _, part := range parts {
			ip := strings.TrimSpace(part)
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		if net.ParseIP(realIP) != nil {
			return realIP
		}
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		if net.ParseIP(host) != nil {
			return host
		}
	}

	return r.RemoteAddr
}

func wantsJSON(r *http.Request) bool {
	if strings.HasPrefix(r.URL.Path, "/api") {
		return true
	}

	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	parts := strings.Split(host, ".")
	if parts[0] == "api" || parts[len(parts)-1] == "api" {
		return true
	}

	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	return strings.Contains(contentType, "json")
}

func hasRDAP(info rdap.Info) bool {
	return info.Country != "" || info.Handle != "" || info.IPVersion != "" || info.Name != "" || info.Type != "" || len(info.Events) > 0
}

// Store defines the methods needed for caching and counting.
type Store interface {
	GetCached(ctx context.Context, ip string) (rdap.Info, time.Time, bool, error)
	SetCached(ctx context.Context, ip string, info rdap.Info, fetchedAt time.Time) error
	IncrementCount(ctx context.Context, ip string) (int64, error)
}

// NewService wires RDAP fetching and cache storage together.
func NewService(store Store, rdapClient RDAPLookup, errorHandler func(error)) *ServiceImpl {
	if errorHandler == nil {
		errorHandler = func(error) {}
	}
	return &ServiceImpl{store: store, rdapClient: rdapClient, onError: errorHandler}
}

// RDAPLookup defines a minimal RDAP client.
type RDAPLookup interface {
	Lookup(ctx context.Context, ip string) (rdap.Info, error)
}

// ServiceImpl is the default implementation of Service.
type ServiceImpl struct {
	store      Store
	rdapClient RDAPLookup
	onError    func(error)
}

// Fetch returns the response for a given IP.
func (s *ServiceImpl) Fetch(ctx context.Context, ip string) (Response, error) {
	var fetchError error
	count, err := s.store.IncrementCount(ctx, ip)
	if err != nil {
		s.OnError(err)
		fetchError = err
	}

	info := rdap.Info{}
	if s.rdapClient != nil {
		cached, fetchedAt, ok, err := s.store.GetCached(ctx, ip)
		if err != nil {
			s.OnError(err)
		}

		if ok && !store.NeedsRefresh(fetchedAt) {
			info = cached
		} else {
			fetched, err := s.rdapClient.Lookup(ctx, ip)
			if err != nil {
				s.OnError(fmt.Errorf("rdap lookup: %w", err))
				if ok {
					info = cached
				}
			} else {
				info = fetched
				if err := s.store.SetCached(ctx, ip, fetched, time.Now().UTC()); err != nil {
					s.OnError(err)
				}
			}
		}
	}

	return Response{IP: ip, CountCall: count, RDAP: info, Events: info.Events, Error: fetchError}, nil
}

// OnError calls the error handler.
func (s *ServiceImpl) OnError(err error) {
	if s.onError != nil {
		s.onError(err)
	}
}
