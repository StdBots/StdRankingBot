package database

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

type userCacheEntry struct {
	LastMsgTime time.Time
	LastMsgHash string
}

// AntiFarmCache handles in-memory rate limiting and spam filtering
type AntiFarmCache struct {
	mu      sync.RWMutex
	entries map[string]userCacheEntry // key: "chatID:userID"
}

// NewAntiFarmCache initializes the cache with background cleanup
func NewAntiFarmCache() *AntiFarmCache {
	c := &AntiFarmCache{
		entries: make(map[string]userCacheEntry),
	}

	// Periodic purge of entries older than 10 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			c.cleanup()
		}
	}()

	return c
}

// CheckAndRecord evaluates if a message is eligible for XP and counts
func (c *AntiFarmCache) CheckAndRecord(chatID, userID int64, text string, cooldownSeconds int, minLen int) (bool, string) {
	if len(text) < minLen {
		return false, "Message too short"
	}

	key := fmt.Sprintf("%d:%d", chatID, userID)
	textHash := fmt.Sprintf("%x", sha256.Sum256([]byte(text)))
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if exists {
		// Cooldown check
		if now.Sub(entry.LastMsgTime) < time.Duration(cooldownSeconds)*time.Second {
			return false, "Cooldown active"
		}

		// Duplicate text check within 60s
		if entry.LastMsgHash == textHash && now.Sub(entry.LastMsgTime) < 60*time.Second {
			return false, "Duplicate message"
		}
	}

	// Update cache
	c.entries[key] = userCacheEntry{
		LastMsgTime: now,
		LastMsgHash: textHash,
	}

	return true, "OK"
}

func (c *AntiFarmCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	for k, v := range c.entries {
		if v.LastMsgTime.Before(cutoff) {
			delete(c.entries, k)
		}
	}
}
