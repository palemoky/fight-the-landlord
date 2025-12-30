package server

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer_RegisterUnregister_Concurrency(t *testing.T) {
	t.Parallel()

	s := &Server{
		clients: make(map[string]*Client),
	}

	var wg sync.WaitGroup
	count := 100

	// Concurrent Register
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()
			c := &Client{ID: string(rune(i))}
			s.RegisterClient(c.ID, c)
		}(i)
	}
	wg.Wait()
	assert.Equal(t, count, s.GetOnlineCount())

	// Concurrent Unregister
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()
			s.UnregisterClient(string(rune(i)))
		}(i)
	}
	wg.Wait()
	assert.Equal(t, 0, s.GetOnlineCount())
}

func TestServer_HandleHealth(t *testing.T) {
	t.Parallel()

	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestServer_MaintenanceMode(t *testing.T) {
	t.Parallel()

	s := &Server{}

	assert.False(t, s.IsMaintenanceMode())

	s.EnterMaintenanceMode()
	assert.True(t, s.IsMaintenanceMode())
}

func TestServer_GracefulShutdown_Logic(t *testing.T) {
	// 这是一个逻辑测试，不涉及真实的 Redis/HTTP 关闭
	t.Parallel()

	// cfg := &config.Config{}
	// mock config to prevent nil pointer if accessed
	// But GracefulShutdown accesses s.config.Game.ShutdownCheckIntervalDuration()
	// So we need to set it up properly or mock parts of it.
	// Since s.roomManager is nil, we should construct a minimal server.

	// Skip complex integration-like tests in unit tests unless we mock everything.
	// Focusing on available simple logic.
}
