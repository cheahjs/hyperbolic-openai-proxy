package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cheahjs/hyperbolic-openai-proxy/internal/api"
	"github.com/cheahjs/hyperbolic-openai-proxy/internal/cache"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/cheahjs/hyperbolic-openai-proxy/internal/api"
	"github.com/cheahjs/hyperbolic-openai-proxy/internal/cache"
)

var (
	baseURL        string
	maxStoreSizeMB int
)

func main() {
	// Resolve environment variables at startup
	expiryDuration := 30 * time.Minute // Default expiry time
	if expiryStr := os.Getenv("IMAGE_EXPIRY"); expiryStr != "" {
		expiryDuration, _ = time.ParseDuration(expiryStr)
	}

	baseURL = os.Getenv("BASE_URL")
	if baseURL != "" && !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}

	maxStoreSizeMB = 0
	if maxStoreSizeStr := os.Getenv("MAX_IMAGE_STORE_SIZE_MB"); maxStoreSizeStr != "" {
		maxStoreSizeMB, _ = strconv.Atoi(maxStoreSizeStr)
	}

	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Log to console by default
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		level, err := zerolog.ParseLevel(envLogLevel)
		if err != nil {
			log.Fatal().Err(err).Msg("Invalid LOG_LEVEL")
		}
		zerolog.SetGlobalLevel(level)
	}

	imageCache := cache.NewImageCache(expiryDuration, maxStoreSizeMB, time.Minute)

	router := api.NewRouter(imageCache, baseURL)

	listenAddr := ":8080"
	if envListenAddr := os.Getenv("LISTEN_ADDR"); envListenAddr != "" {
		listenAddr = envListenAddr
	}

	err := http.ListenAndServe(listenAddr, router)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
