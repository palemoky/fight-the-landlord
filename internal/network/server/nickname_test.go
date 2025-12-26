package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateNickname(t *testing.T) {
	// Simple test to ensure not empty
	name1 := GenerateNickname()
	assert.NotEmpty(t, name1)

	name2 := GenerateNickname()
	assert.NotEmpty(t, name2)
	// It's possible they are the same due to randomness, but highly unlikely if pool is large enough.
	// We won't assert inequality to avoid flaky tests, but we checked basic generation.
}
