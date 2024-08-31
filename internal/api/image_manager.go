package api

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var (
	ErrImageNotFound = errors.New("image not found")
	ErrImageExpired  = errors.New("image expired")
)

type ImageManager struct {
	basePath        string
	expiryDuration  time.Duration
	maxStoreSizeMB  int
	store           map[string]imageEntry
	mu              sync.Mutex
	cleanupInterval time.Duration
}

type imageEntry struct {
	data      []byte
	expiresAt time.Time
}

func NewImageManager(basePath string, expiryDuration time.Duration, maxStoreSizeMB int, cleanupInterval time.Duration) (*ImageManager, error) {
	if basePath != "" {
		if err := os.MkdirAll(basePath, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create image store directory: %w", err)
		}
	}

	manager := &ImageManager{
		basePath:        basePath,
		expiryDuration:  expiryDuration,
		maxStoreSizeMB:  maxStoreSizeMB,
		store:           make(map[string]imageEntry),
		cleanupInterval: cleanupInterval,
	}

	if basePath == "" {
		go func() {
			ticker := time.NewTicker(manager.cleanupInterval)
			defer ticker.Stop()

			for range ticker.C {
				manager.cleanup()
			}
		}()
	}

	return manager, nil
}

func (manager *ImageManager) StoreImageWithPrompt(prompt string, imageData []byte) (string, error) {
	decodedImage, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	id := uuid.New().String()

	if manager.basePath != "" {
		imagePath := filepath.Join(manager.basePath, id+"."+format)
		imageFile, err := os.Create(imagePath)
		if err != nil {
			return "", fmt.Errorf("failed to create image file: %w", err)
		}
		defer imageFile.Close()

		if err := png.Encode(imageFile, decodedImage); err != nil {
			return "", fmt.Errorf("failed to encode image: %w", err)
		}

		promptPath := filepath.Join(manager.basePath, id+".txt")
		if err := ioutil.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
			return "", fmt.Errorf("failed to write prompt file: %w", err)
		}

		log.Info().Str("path", imagePath).Msg("Stored image")

		return imagePath, nil
	} else {
		manager.mu.Lock()
		defer manager.mu.Unlock()

		expiresAt := time.Now().Add(manager.expiryDuration)
		manager.store[id] = imageEntry{imageData, expiresAt}

		log.Info().Str("id", id).Msg("Stored image in cache")

		return id, nil
	}
}

func (manager *ImageManager) GetImage(id string) ([]byte, error) {
	if manager.basePath != "" {
		imagePath := filepath.Join(manager.basePath, id)

		// Find the image file with the matching ID and a supported extension
		matches, err := filepath.Glob(imagePath + ".*")
		if err != nil {
			log.Error().Err(err).Str("id", id).Msg("Failed to find image file")
			return nil, ErrImageNotFound
		}

		if len(matches) > 0 {
			imageData, err := ioutil.ReadFile(matches[0])
			if err != nil {
				log.Error().Err(err).Str("id", id).Msg("Failed to read image file")
				return nil, fmt.Errorf("failed to read image file: %w", err)
			}
			return imageData, nil
		}

		return nil, ErrImageNotFound
	} else {
		manager.mu.Lock()
		defer manager.mu.Unlock()

		entry, ok := manager.store[id]
		if !ok {
			log.Warn().Str("id", id).Msg("Image not found in cache")
			return nil, ErrImageNotFound
		}

		if time.Now().After(entry.expiresAt) {
			log.Warn().Str("id", id).Msg("Image expired in cache")
			delete(manager.store, id)
			return nil, ErrImageExpired
		}

		log.Info().Str("id", id).Msg("Retrieved image from cache")

		return entry.data, nil
	}
}

func generateUniqueID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (manager *ImageManager) cleanup() {
	log.Debug().Msg("Cleaning up image cache")
	manager.mu.Lock()
	defer manager.mu.Unlock()

	now := time.Now()
	for id, entry := range manager.store {
		if now.After(entry.expiresAt) {
			delete(manager.store, id)
			log.Debug().Str("id", id).Msg("Removed expired image from cache")
		}
	}
}
