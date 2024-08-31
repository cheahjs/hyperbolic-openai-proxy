package api

import (
	"encoding/base64"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/cheahjs/hyperbolic-openai-proxy/internal/cache"
	"github.com/gorilla/mux"
)

func (router *Router) imageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	imageData, err := router.imageCache.GetImage(id)
	if err == cache.ErrImageNotFound {
		log.Error().Err(err).Str("id", id).Msg("Image not found")
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	} else if err == cache.ErrImageExpired {
		log.Error().Err(err).Str("id", id).Msg("Image expired")
		http.Error(w, "Image expired", http.StatusGone)
		return
	} else if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to retrieve image")
		http.Error(w, "Failed to retrieve image", http.StatusInternalServerError)
		return
	}

	decodedImage, err := base64.StdEncoding.DecodeString(string(imageData))
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode image")
		http.Error(w, "Failed to decode image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(decodedImage)
}
