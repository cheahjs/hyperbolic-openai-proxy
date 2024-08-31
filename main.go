package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// OpenAIRequest represents an OpenAI image generation request
type OpenAIRequest struct {
	Model          string  `json:"model"`
	Prompt         string  `json:"prompt"`
	N              *int    `json:"n,omitempty"`
	Size           *string `json:"size,omitempty"`
	ResponseFormat *string `json:"response_format,omitempty"`
}

// HyperbolicRequest represents a Hyperbolic image generation request
type HyperbolicRequest struct {
	ModelName string `json:"model_name"`
	Prompt    string `json:"prompt"`
	Height    int    `json:"height"`
	Width     int    `json:"width"`
	Backend   string `json:"backend"`
}

// HyperbolicResponse represents a Hyperbolic API response
type HyperbolicResponse struct {
	Images        []HyperbolicImage `json:"images"`
	InferenceTime float64           `json:"inference_time"`
}

// HyperbolicImage represents a generated image in the Hyperbolic API response
type HyperbolicImage struct {
	Index      int    `json:"index"`
	Image      string `json:"image"`
	RandomSeed int64  `json:"random_seed"`
}

// imageEntry stores image data and its expiration time
type imageEntry struct {
	data      []byte
	expiresAt time.Time
}

var imageStore = make(map[string]imageEntry)

var expiryDuration time.Duration
var maxStoreSizeMB int

// OpenAIResponse represents an OpenAI API response
type OpenAIResponse struct {
	Created int64         `json:"created"`
	Data    []OpenAIImage `json:"data"`
}

// OpenAIImage represents a generated image in the OpenAI API response
type OpenAIImage struct {
	B64JSON       string `json:"b64_json,omitempty"`
	URL           string `json:"url,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

func convertRequest(openAIRequest OpenAIRequest) HyperbolicRequest {
	var hyperbolicRequest HyperbolicRequest

	hyperbolicRequest.ModelName = openAIRequest.Model
	hyperbolicRequest.Prompt = openAIRequest.Prompt

	if openAIRequest.Size != nil {
		var sizeSplit = strings.Split(*openAIRequest.Size, "x")
		if len(sizeSplit) != 2 {
			log.Println("Invalid size")
			return hyperbolicRequest
		}
		height, err := strconv.Atoi(sizeSplit[0])
		if err != nil {
			log.Println(err)
			return hyperbolicRequest
		}
		width, err := strconv.Atoi(sizeSplit[1])
		if err != nil {
			log.Println(err)
			return hyperbolicRequest
		}
		hyperbolicRequest.Height = height
		hyperbolicRequest.Width = width
	} else {
		hyperbolicRequest.Height = 1024
		hyperbolicRequest.Width = 1024
	}

	hyperbolicRequest.Backend = "auto"

	return hyperbolicRequest
}

func convertResponse(hyperbolicResponse HyperbolicResponse, openAIRequest OpenAIRequest, baseURL string) OpenAIResponse {
	var openAIResponse OpenAIResponse

	openAIResponse.Created = time.Now().Unix()

	for _, image := range hyperbolicResponse.Images {
		var openAIImage OpenAIImage
		if openAIRequest.ResponseFormat != nil && *openAIRequest.ResponseFormat == "url" {
			id := generateUniqueID()

			expiresAt := time.Now().Add(expiryDuration)

			imageStore[id] = imageEntry{[]byte(image.Image), expiresAt}
			openAIImage.URL = fmt.Sprintf("%s/images/%s", baseURL, id)

		} else {
			openAIImage.B64JSON = image.Image
		}
		openAIResponse.Data = append(openAIResponse.Data, openAIImage)
	}

	return openAIResponse
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	entry, ok := imageStore[id]
	if !ok {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	if time.Now().After(entry.expiresAt) {
		delete(imageStore, id)
		http.Error(w, "Image expired", http.StatusGone)
		return
	}

	decodedImage, err := base64.StdEncoding.DecodeString(string(entry.data))
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to decode image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(decodedImage)
}

func imageGenerationHandler(w http.ResponseWriter, r *http.Request) {
	var openAIRequest OpenAIRequest

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &openAIRequest)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if openAIRequest.N != nil && *openAIRequest.N != 1 {
		http.Error(w, "n must be 1", http.StatusBadRequest)
		return
	}

	hyperbolicRequest := convertRequest(openAIRequest)

	jsonBody, err := json.Marshal(hyperbolicRequest)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST", "https://api.hyperbolic.xyz/v1/image/generation", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Pass through headers from the original request
	for key, value := range r.Header {
		if key != "Host" {
			req.Header.Set(key, value[0])
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var hyperbolicResponse HyperbolicResponse
	err = json.Unmarshal(body, &hyperbolicResponse)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://" + r.Host
	}

	openAIResponse := convertResponse(hyperbolicResponse, openAIRequest, baseURL)

	jsonBody, err = json.Marshal(openAIResponse)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBody)
}

func main() {
	// Resolve environment variables at startup
	expiryDuration = 30 * time.Minute // Default expiry time
	if expiryStr := os.Getenv("IMAGE_EXPIRY"); expiryStr != "" {
		expiryDuration, _ = time.ParseDuration(expiryStr)
	}

	maxStoreSizeMB = 0
	if maxStoreSizeStr := os.Getenv("MAX_IMAGE_STORE_SIZE_MB"); maxStoreSizeStr != "" {
		maxStoreSizeMB, _ = strconv.Atoi(maxStoreSizeStr)
	}

	r := mux.NewRouter()
	r.HandleFunc("/image/generation", imageGenerationHandler).Methods("POST")
	r.HandleFunc("/images/{id}", imageHandler).Methods("GET")

	// Start a goroutine to clean up expired images
	go func() {
		for {
			time.Sleep(time.Minute)
			cleanupImageStore()
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", r))
}

func generateUniqueID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func cleanupImageStore() {
	now := time.Now()
	totalSizeMB := 0
	for id, entry := range imageStore {
		if now.After(entry.expiresAt) {
			delete(imageStore, id)
		} else if maxStoreSizeMB > 0 {
			totalSizeMB += len(entry.data) / (1024 * 1024)
			if totalSizeMB > maxStoreSizeMB {
				delete(imageStore, id)
			}
		}
	}
}
