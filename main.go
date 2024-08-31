package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cheahjs/hyperbolic-openai-proxy/internal/api"
	"github.com/cheahjs/hyperbolic-openai-proxy/internal/cache"
)

var (
	baseURL        string
	maxStoreSizeMB int
)

func cleanupImageStore() {
	// TODO: Implement image store cleanup logic
}

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

	imageCache := cache.NewImageCache(expiryDuration, maxStoreSizeMB, time.Minute)

	router := api.NewRouter(imageCache, baseURL)


	listenAddr := ":8080"
	if envListenAddr := os.Getenv("LISTEN_ADDR"); envListenAddr != "" {
		listenAddr = envListenAddr
	}

	err := http.ListenAndServe(listenAddr, router)
	if err != nil {
		log.Fatal(err)
	}
}
