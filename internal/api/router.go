package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/cheahjs/hyperbolic-openai-proxy/internal/cache"
	"github.com/gorilla/mux"
)

type Router struct {
	router     *mux.Router
	imageCache *cache.ImageCache
	baseURL    string
}

func NewRouter(imageCache *cache.ImageCache, baseURL string) *Router {
	r := mux.NewRouter()
	router := &Router{
		router:     r,
		imageCache: imageCache,
		baseURL:    baseURL,
	}

	r.HandleFunc("/image/generation", router.imageGenerationHandler).Methods("POST")
	r.HandleFunc("/images/{id}", router.imageHandler).Methods("GET")

	r.Use(loggingMiddleware)

	return router
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("Request started")

		next.ServeHTTP(w, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Dur("duration", time.Since(start)).
			Msg("Request completed")
	})
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.router.ServeHTTP(w, r)
}

func (router *Router) getBaseUrl(r *http.Request) string {
	if router.baseURL != "" {
		return router.baseURL
	}
	return "http://" + r.Host
}

func (router *Router) imageGenerationHandler(w http.ResponseWriter, r *http.Request) {
	var openAIRequest OpenAIRequest

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Info().Str("model", openAIRequest.Model).Msg("Model parsed")

	err = json.Unmarshal(body, &openAIRequest)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if openAIRequest.Model == "" {
		http.Error(w, "model is required", http.StatusBadRequest)
		return
	}

	if openAIRequest.N != nil && *openAIRequest.N != 1 {
		http.Error(w, "n must be 1", http.StatusBadRequest)
		return
	}

	hyperbolicRequest, err := convertRequest(&openAIRequest)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonBody, err := json.Marshal(hyperbolicRequest)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal request body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST", "https://api.hyperbolic.xyz/v1/image/generation", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Pass through headers from the original request
	for key, value := range r.Header {
		if key != "Host" {
			req.Header[key] = value
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Error().Err(err).Str("status", resp.Status).Str("body", string(body)).Msg("Failed to send request")
		http.Error(w, "Failed to generate image", resp.StatusCode)
		return
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read response body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var hyperbolicResponse HyperbolicResponse
	err = json.Unmarshal(body, &hyperbolicResponse)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal response body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	openAIResponse, err := convertResponse(hyperbolicResponse, openAIRequest, router.getBaseUrl(r), router.imageCache)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert response")
		http.Error(w, "Failed to convert response", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, openAIResponse)
}
