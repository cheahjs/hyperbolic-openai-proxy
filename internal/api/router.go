package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

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

	return router
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
			req.Header[key] = value
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

	openAIResponse, err := convertResponse(hyperbolicResponse, openAIRequest, router.getBaseUrl(r), router.imageCache)
	if err != nil {
		log.Println("Error converting response:", err)
		http.Error(w, "Failed to convert response", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, openAIResponse)
}

