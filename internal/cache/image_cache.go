package cache

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var ErrImageNotFound = errors.New("image not found")
var ErrImageExpired = errors.New("image expired")

type ImageCache struct {
	store          map[string]imageEntry
	expiryDuration time.Duration
	maxStoreSizeMB int
	mu             sync.Mutex
}

type imageEntry struct {
	data      []byte
	expiresAt time.Time
}

func NewImageCache(expiryDuration time.Duration, maxStoreSizeMB int) *ImageCache {
	return &ImageCache{
		store:          make(map[string]imageEntry),
		expiryDuration: expiryDuration,
		maxStoreSizeMB: maxStoreSizeMB,
	}
}

func (c *ImageCache) StoreImage(data []byte) (string, error) {
	c.mu.Lock()         // Acquire lock before accessing the map
	defer c.mu.Unlock() // Release lock when done

	id, err := generateUniqueID()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(c.expiryDuration)
	c.store[id] = imageEntry{data, expiresAt}

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
	c.mu.Lock()         // Acquire lock before accessing the map
	defer c.mu.Unlock() // Release lock when done

	entry, ok := c.store[id]
	if !ok {
		return nil, ErrImageNotFound
	}

	if time.Now().After(entry.expiresAt) {
		delete(c.store, id)
		return nil, ErrImageExpired
	}

	return entry.data, nil
}
