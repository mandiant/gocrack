package notifications

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheLastPasswordSent(t *testing.T) {
	cache := NewCache(time.Duration(1 * time.Second))
	defer cache.Stop()

	canSend := cache.CanSendEmail("my-task-id")
	assert.True(t, canSend)

	canSend = cache.CanSendEmail("my-task-id")
	assert.False(t, canSend)

	time.Sleep(2 * time.Second)

	canSend = cache.CanSendEmail("my-task-id")
	assert.True(t, canSend)
}
