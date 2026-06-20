package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinHelper(t *testing.T) {
	assert.Equal(t, "nextjs, nestjs", join([]string{"nextjs", "nestjs"}, ", "))
	assert.Equal(t, "none", join([]string{}, ", "))
	assert.Equal(t, "single", join([]string{"single"}, ", "))
}

func TestContainsHelper(t *testing.T) {
	assert.True(t, contains([]string{"nextjs", "nestjs"}, "nextjs"))
	assert.True(t, contains([]string{"nextjs", "nestjs"}, "NEXTJS"))
	assert.False(t, contains([]string{"nextjs"}, "fastapi"))
}
