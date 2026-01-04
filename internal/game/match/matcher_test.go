package match

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palemoky/fight-the-landlord/internal/testutil"
)

func TestMatcher_QueueOps(t *testing.T) {
	// As long as we keep queue size < 3, it won't call CreateRoom.
	matcher := NewMatcher(MatcherDeps{}) // nil dependencies for testing

	c1 := &testutil.SimpleClient{ID: "p1", Name: "Player1"}
	c2 := &testutil.SimpleClient{ID: "p2", Name: "Player2"}

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
