package api

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/cheahjs/hyperbolic-openai-proxy/internal/cache"
)

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

func convertResponse(hyperbolicResponse HyperbolicResponse, openAIRequest OpenAIRequest, baseURL string, imageCache *cache.ImageCache) (OpenAIResponse, error) {
	var openAIResponse OpenAIResponse

	openAIResponse.Created = time.Now().Unix()

	for _, image := range hyperbolicResponse.Images {
		var openAIImage OpenAIImage
		if openAIRequest.ResponseFormat != nil && *openAIRequest.ResponseFormat == "b64_json" {
			openAIImage.B64JSON = image.Image
		} else { // default to URL
			id, err := imageCache.StoreImage([]byte(image.Image))
			if err != nil {
				return openAIResponse, fmt.Errorf("failed to store image: %w", err)
			}
			openAIImage.URL = fmt.Sprintf("%s/images/%s", baseURL, id)
		}
		openAIResponse.Data = append(openAIResponse.Data, openAIImage)
	}

	return openAIResponse, nil
}

