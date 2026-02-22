package pihole

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDisableBlocking_HappyPath(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/auth":
			var req authRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode auth request: %v", err)
			}
			if req.Password != "testpass" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			json.NewEncoder(w).Encode(authResponse{
				Session: struct {
					SID string `json:"sid"`
				}{SID: "test-sid-123"},
			})

		case r.Method == http.MethodPost && r.URL.Path == "/api/dns/blocking":
			if r.Header.Get("X-FTL-SID") != "test-sid-123" {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			var req blockingRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode blocking request: %v", err)
			}
			if req.Blocking != false || req.Timer != 1800 {
				t.Errorf("unexpected blocking request: %+v", req)
			}
			w.WriteHeader(http.StatusOK)

		case r.Method == http.MethodDelete && r.URL.Path == "/api/auth":
			if r.Header.Get("X-FTL-SID") != "test-sid-123" {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestClient(srv)

	err := client.DisableBlocking(context.Background(), "testpass", 1800)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDisableBlocking_AuthFailure(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/auth" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	client := newTestClient(srv)

	err := client.DisableBlocking(context.Background(), "wrongpass", 1800)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "authentication") {
		t.Errorf("expected authentication error, got: %v", err)
	}
}

func TestDisableBlocking_BlockingFailure(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/auth":
			json.NewEncoder(w).Encode(authResponse{
				Session: struct {
					SID string `json:"sid"`
				}{SID: "test-sid"},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/dns/blocking":
			http.Error(w, "internal server error", http.StatusInternalServerError)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/auth":
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestClient(srv)

	err := client.DisableBlocking(context.Background(), "testpass", 1800)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "disable blocking") {
		t.Errorf("expected disable blocking error, got: %v", err)
	}
}

func TestDisableBlocking_InvalidJSON(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/auth" {
			w.Write([]byte("not json"))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	client := newTestClient(srv)

	err := client.DisableBlocking(context.Background(), "testpass", 1800)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "authentication") {
		t.Errorf("expected authentication error, got: %v", err)
	}
}

func TestEnableBlocking_HappyPath(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/auth":
			json.NewEncoder(w).Encode(authResponse{
				Session: struct {
					SID string `json:"sid"`
				}{SID: "test-sid-123"},
			})

		case r.Method == http.MethodGet && r.URL.Path == "/api/dns/blocking":
			if r.Header.Get("X-FTL-SID") != "test-sid-123" {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"blocking": "disabled"})

		case r.Method == http.MethodPost && r.URL.Path == "/api/dns/blocking":
			if r.Header.Get("X-FTL-SID") != "test-sid-123" {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			var req blockingRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode blocking request: %v", err)
			}
			if req.Blocking != true {
				t.Errorf("expected blocking=true, got %v", req.Blocking)
			}
			w.WriteHeader(http.StatusOK)

		case r.Method == http.MethodDelete && r.URL.Path == "/api/auth":
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestClient(srv)

	err := client.EnableBlocking(context.Background(), "testpass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnableBlocking_AlreadyEnabled(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/auth":
			json.NewEncoder(w).Encode(authResponse{
				Session: struct {
					SID string `json:"sid"`
				}{SID: "test-sid-123"},
			})

		case r.Method == http.MethodGet && r.URL.Path == "/api/dns/blocking":
			json.NewEncoder(w).Encode(map[string]string{"blocking": "enabled"})

		case r.Method == http.MethodDelete && r.URL.Path == "/api/auth":
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestClient(srv)

	err := client.EnableBlocking(context.Background(), "testpass")
	if !errors.Is(err, ErrAlreadyEnabled) {
		t.Fatalf("expected ErrAlreadyEnabled, got: %v", err)
	}
}

func TestEnableBlocking_AuthFailure(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/auth" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	client := newTestClient(srv)

	err := client.EnableBlocking(context.Background(), "wrongpass")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "authentication") {
		t.Errorf("expected authentication error, got: %v", err)
	}
}

func TestEnableBlocking_StatusCheckFailure(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/auth":
			json.NewEncoder(w).Encode(authResponse{
				Session: struct {
					SID string `json:"sid"`
				}{SID: "test-sid-123"},
			})

		case r.Method == http.MethodGet && r.URL.Path == "/api/dns/blocking":
			http.Error(w, "internal server error", http.StatusInternalServerError)

		case r.Method == http.MethodDelete && r.URL.Path == "/api/auth":
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestClient(srv)

	err := client.EnableBlocking(context.Background(), "testpass")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "check status") {
		t.Errorf("expected check status error, got: %v", err)
	}
}

func newTestClient(srv *httptest.Server) *Client {
	// Strip the https:// prefix to get just the host
	host := strings.TrimPrefix(srv.URL, "https://")
	c := NewClient(host)
	c.httpClient = srv.Client()
	return c
}
