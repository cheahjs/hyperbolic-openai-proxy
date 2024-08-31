package api

import (
	"encoding/base64"
	"log"
	"net/http"

	"github.com/cheahjs/hyperbolic-openai-proxy/internal/cache"
	"github.com/gorilla/mux"
)

func (router *Router) imageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	imageData, err := router.imageCache.GetImage(id)
	if err == cache.ErrImageNotFound {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	} else if err == cache.ErrImageExpired {
		http.Error(w, "Image expired", http.StatusGone)
		return
	} else if err != nil {
		log.Println(err)
		http.Error(w, "Failed to retrieve image", http.StatusInternalServerError)
		return
	}

	decodedImage, err := base64.StdEncoding.DecodeString(string(imageData))
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to decode image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(decodedImage)
}
