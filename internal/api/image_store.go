package api

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type ImageStore struct {
	basePath string
}

func NewImageStore(basePath string) (*ImageStore, error) {
	if basePath == "" {
		return nil, nil // Image store is disabled
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create image store directory: %w", err)
	}

	return &ImageStore{basePath: basePath}, nil
}

func (store *ImageStore) StoreImageWithPrompt(prompt string, imageData []byte) (string, error) {
	decodedImage, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	id := uuid.New().String()
	imagePath := filepath.Join(store.basePath, id+"."+format)
	imageFile, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to create image file: %w", err)
	}
	defer imageFile.Close()

	if err := png.Encode(imageFile, decodedImage); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	promptPath := filepath.Join(store.basePath, id+".txt")
	if err := ioutil.WriteFile(promptPath, []byte(prompt), 0644); err != nil {
		return "", fmt.Errorf("failed to write prompt file: %w", err)
	}

	log.Info().Str("path", imagePath).Msg("Stored image")

	return imagePath, nil
}

func (store *ImageStore) GetImagePath(id string) string {
	imagePath := filepath.Join(store.basePath, id)

	// Find the image file with the matching ID and a supported extension
	matches, err := filepath.Glob(imagePath + ".*")
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to find image file")
		return ""
	}

	if len(matches) > 0 {
		return matches[0]
	}

	return ""
}
