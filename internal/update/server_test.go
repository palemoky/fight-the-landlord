package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequiresUpgrade(t *testing.T) {
	tests := []struct {
		name    string
		req     *ServerRequirement
		current string
		want    bool
	}{
		{"no minimum set", &ServerRequirement{MinClientVersion: ""}, "v1.0.0", false},
		{"current below minimum", &ServerRequirement{MinClientVersion: "v1.2.0"}, "v1.1.0", true},
		{"current equals minimum", &ServerRequirement{MinClientVersion: "v1.2.0"}, "v1.2.0", false},
		{"current above minimum", &ServerRequirement{MinClientVersion: "v1.2.0"}, "v1.3.0", false},
		{"nil requirement", nil, "v1.0.0", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.req.RequiresUpgrade(tt.current); got != tt.want {
				t.Errorf("RequiresUpgrade(%q) = %v, want %v", tt.current, got, tt.want)
			}
		})
	}
}

func TestVersionEndpoint(t *testing.T) {
	tests := []struct {
		serverURL string
		want      string
		wantErr   bool
	}{
		{"ws://localhost:1780/ws", "http://localhost:1780/version", false},
		{"wss://example.com/ws", "https://example.com/version", false},
		{"http://localhost:1780/ws", "http://localhost:1780/version", false},
		{"ws://host:1780/ws?token=abc", "http://host:1780/version", false},
		{"ftp://host/ws", "", true},
	}
	for _, tt := range tests {
		got, err := versionEndpoint(tt.serverURL)
		if tt.wantErr {
			if err == nil {
				t.Errorf("versionEndpoint(%q) expected error, got %q", tt.serverURL, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("versionEndpoint(%q) unexpected error: %v", tt.serverURL, err)
		}
		if got != tt.want {
			t.Errorf("versionEndpoint(%q) = %q, want %q", tt.serverURL, got, tt.want)
		}
	}
}

func TestFetchServerRequirement(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/version" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"server_version":"v1.2.0","min_client_version":"v1.1.0"}`))
	}))
	defer srv.Close()

	// httptest 提供 http:// 地址，转换为 ws:// 以模拟客户端连接地址。
	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1) + "/ws"

	req, err := FetchServerRequirement(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("FetchServerRequirement returned error: %v", err)
	}
	if req.ServerVersion != "v1.2.0" {
		t.Errorf("ServerVersion = %q, want v1.2.0", req.ServerVersion)
	}
	if req.MinClientVersion != "v1.1.0" {
		t.Errorf("MinClientVersion = %q, want v1.1.0", req.MinClientVersion)
	}
	if !req.RequiresUpgrade("v1.0.0") {
		t.Errorf("expected v1.0.0 to require upgrade")
	}
	if req.RequiresUpgrade("v1.1.0") {
		t.Errorf("expected v1.1.0 to satisfy requirement")
	}
}
