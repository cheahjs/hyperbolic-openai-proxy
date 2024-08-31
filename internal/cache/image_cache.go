package cache

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	ErrImageNotFound = errors.New("image not found")
	ErrImageExpired  = errors.New("image expired")
)

type ImageCache struct {
	store           map[string]imageEntry
	expiryDuration  time.Duration
	maxStoreSizeMB  int
	mu              sync.Mutex
	cleanupInterval time.Duration
}

type imageEntry struct {
	data      []byte
	expiresAt time.Time
}

func NewImageCache(expiryDuration time.Duration, maxStoreSizeMB int, cleanupInterval time.Duration) *ImageCache {
	c := &ImageCache{
		store:           make(map[string]imageEntry),
		expiryDuration:  expiryDuration,
		maxStoreSizeMB:  maxStoreSizeMB,
		cleanupInterval: cleanupInterval,
	}

	go func() {
		ticker := time.NewTicker(c.cleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			c.cleanup()
		}
	}()

	return c
}

func (c *ImageCache) StoreImage(data []byte) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id, err := generateUniqueID()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate unique ID")
		return "", err
	}

	expiresAt := time.Now().Add(c.expiryDuration)
	c.store[id] = imageEntry{data, expiresAt}

	log.Info().Str("id", id).Msg("Stored image in cache")

	return id, nil
}

func (c *ImageCache) StoreImageWithPrompt(prompt string, data []byte) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id, err := generateUniqueID()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate unique ID")
		return "", err
	}

	expiresAt := time.Now().Add(c.expiryDuration)
	c.store[id] = imageEntry{data, expiresAt}

	log.Info().Str("id", id).Msg("Stored image in cache")

	return id, nil
}

func generateUniqueID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (c *ImageCache) GetImage(id string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.store[id]
	if !ok {
		log.Warn().Str("id", id).Msg("Image not found in cache")
		return nil, ErrImageNotFound
	}

	if time.Now().After(entry.expiresAt) {
		log.Warn().Str("id", id).Msg("Image expired in cache")
		delete(c.store, id)
		return nil, ErrImageExpired
	}

	log.Info().Str("id", id).Msg("Retrieved image from cache")

	return entry.data, nil
}

func (c *ImageCache) cleanup() {
	log.Debug().Msg("Cleaning up image cache")
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for id, entry := range c.store {
		if now.After(entry.expiresAt) {
			delete(c.store, id)
			log.Debug().Str("id", id).Msg("Removed expired image from cache")
		}
	}
}
