package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int // -1: a<b, 0: a==b, 1: a>b
	}{
		{"equal with v prefix", "v1.2.3", "v1.2.3", 0},
		{"equal mixed prefix", "1.2.3", "v1.2.3", 0},
		{"patch newer", "v1.2.3", "v1.2.4", -1},
		{"patch older", "v1.2.4", "v1.2.3", 1},
		{"minor newer", "v1.2.9", "v1.3.0", -1},
		{"major newer", "v1.9.9", "v2.0.0", -1},
		{"missing patch treated as zero", "v1.2", "v1.2.0", 0},
		{"prerelease older than release", "v1.2.3-rc.1", "v1.2.3", -1},
		{"release newer than prerelease", "v1.2.3", "v1.2.3-rc.1", 1},
		{"prerelease ordering", "v1.2.3-rc.1", "v1.2.3-rc.2", -1},
		{"dev compares as zero base", "dev", "v0.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareVersions(tt.a, tt.b)
			if sign(got) != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want sign %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}

func TestIsDevVersion(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"", true},
		{"dev", true},
		{"unknown", true},
		{"  dev  ", true},
		{"v1.0.0", false},
		{"1.0.0", false},
	}

	for _, tt := range tests {
		if got := IsDevVersion(tt.version); got != tt.want {
			t.Errorf("IsDevVersion(%q) = %v, want %v", tt.version, got, tt.want)
		}
	}
}

func TestCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v2.0.0","html_url":"https://example.com/releases/v2.0.0"}`))
	}))
	defer srv.Close()

	res, err := checkAt(context.Background(), "v1.0.0", srv.URL)
	if err != nil {
		t.Fatalf("checkAt returned error: %v", err)
	}
	if !res.HasUpdate {
		t.Errorf("expected HasUpdate=true for v1.0.0 -> v2.0.0")
	}
	if res.LatestVersion != "v2.0.0" {
		t.Errorf("LatestVersion = %q, want v2.0.0", res.LatestVersion)
	}

	res, err = checkAt(context.Background(), "v2.0.0", srv.URL)
	if err != nil {
		t.Fatalf("checkAt returned error: %v", err)
	}
	if res.HasUpdate {
		t.Errorf("expected HasUpdate=false when already on latest")
	}
}

func TestAssetName(t *testing.T) {
	tests := []struct {
		goos   string
		goarch string
		want   string
	}{
		{"linux", "amd64", "fight-the-landlord-linux-amd64"},
		{"darwin", "arm64", "fight-the-landlord-darwin-arm64"},
		{"windows", "amd64", "fight-the-landlord-windows-amd64.exe"},
	}
	for _, tt := range tests {
		if got := AssetName(tt.goos, tt.goarch); got != tt.want {
			t.Errorf("AssetName(%q, %q) = %q, want %q", tt.goos, tt.goarch, got, tt.want)
		}
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("hello fight-the-landlord")
	sum := sha256.Sum256(data)
	hexSum := hex.EncodeToString(sum[:])

	// sha256sum 输出格式：<hex>␠␠<filename>
	good := []byte(hexSum + "  fight-the-landlord-linux-amd64\n")
	if err := verifyChecksum(data, good); err != nil {
		t.Errorf("verifyChecksum with valid sum returned error: %v", err)
	}

	bad := []byte("0000000000000000000000000000000000000000000000000000000000000000  x\n")
	if err := verifyChecksum(data, bad); err == nil {
		t.Errorf("verifyChecksum with wrong sum should fail")
	}

	if err := verifyChecksum(data, []byte("")); err == nil {
		t.Errorf("verifyChecksum with empty file should fail")
	}
}
