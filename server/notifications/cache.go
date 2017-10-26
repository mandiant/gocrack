/* cache.go provides functionality to prevent spamming users with many "password cracked" emails */

package notifications

import (
	"sync"
	"time"
)

// CacheLastPasswordSent stores information about when we last sent password cracked emails for a specific task
type CacheLastPasswordSent struct {
	data   map[string]time.Time // map[taskid (string)]last notification sent (time.Time)
	mu     *sync.Mutex
	expire time.Duration
	stop   chan bool
}

// NewCache creates a password notification sent cache
func NewCache(expireAfter time.Duration) *CacheLastPasswordSent {
	cache := &CacheLastPasswordSent{
		data:   make(map[string]time.Time),
		mu:     &sync.Mutex{},
		expire: expireAfter,
		stop:   make(chan bool, 1),
	}

	go cache.cleanup()
	return cache
}

// cleanup runs periodically to clean up the anti spam cache
func (s *CacheLastPasswordSent) cleanup() {
	tickEvery := time.NewTicker(1 * time.Minute)
	defer tickEvery.Stop()

	clearOutAfter := time.Duration(10 * time.Minute)

Loop:
	for {
		select {
		case <-s.stop:
			break Loop
		case <-tickEvery.C:
		}

		s.mu.Lock()
		now := time.Now().UTC()
		for key, t := range s.data {
			if now.Sub(t) >= clearOutAfter {
				delete(s.data, key)
			}
		}
		s.mu.Unlock()
	}
}

// CanSendEmail returns a boolean indicating if we're able to send an email. If checks to see if the taskid
// exists in a lookup table which stores the last time an email was sent out for the taskid. If a value exists
// in the table and it's greater than the expiration, it will be reset to the current time and return true.
// If the lookup fails, it will set the current time for the task and return true.
func (s *CacheLastPasswordSent) CanSendEmail(taskID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if t, ok := s.data[taskID]; ok {
		// check and see if the last notification we sent is
		// greater than our expire (meaning we can send a new password out)
		if time.Now().UTC().Sub(t) > s.expire {
			goto SetCache
		}
		return false
	}

SetCache:
	s.data[taskID] = time.Now().UTC()
	return true
}

// Stop the password cache expiration goroutine
func (s *CacheLastPasswordSent) Stop() {
	close(s.stop)

	s.mu.Lock()
	s.data = make(map[string]time.Time)
	s.mu.Unlock()
}
