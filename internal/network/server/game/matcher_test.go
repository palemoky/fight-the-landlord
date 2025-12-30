package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatcher_QueueOps(t *testing.T) {
	// Matcher relies on Server only for createMatchRoom.
	// As long as we keep queue size < 3, it won't call CreateRoom.
	matcher := NewMatcher(nil) // nil server should be safe

	c1 := &MockClient{ID: "p1", Name: "Player1"}
	c2 := &MockClient{ID: "p2", Name: "Player2"}

	// Add c1
	matcher.AddToQueue(c1)
	assert.Equal(t, 1, matcher.GetQueueLength())

	// Add c1 again (should be ignored)
	matcher.AddToQueue(c1)
	assert.Equal(t, 1, matcher.GetQueueLength())

	// Add c2
	matcher.AddToQueue(c2)
	assert.Equal(t, 2, matcher.GetQueueLength())

	// Remove c1
	matcher.RemoveFromQueue(c1)
	assert.Equal(t, 1, matcher.GetQueueLength())

	// Remove c1 again (should be no-op)
	matcher.RemoveFromQueue(c1)
	assert.Equal(t, 1, matcher.GetQueueLength())

	// Remove c2
	matcher.RemoveFromQueue(c2)
	assert.Equal(t, 0, matcher.GetQueueLength())
}
