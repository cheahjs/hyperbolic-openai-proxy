package api

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/cheahjs/hyperbolic-openai-proxy/internal/cache"
)

func convertRequest(req *OpenAIRequest) (*HyperbolicRequest, error) {
	var hyperbolicReq HyperbolicRequest

	hyperbolicReq.ModelName = req.Model
	hyperbolicReq.Prompt = req.Prompt

	if req.Size != nil {
		var sizeSplit = strings.Split(*req.Size, "x")
		if len(sizeSplit) != 2 {
			return nil, fmt.Errorf("invalid size: %s", *req.Size)
		}
		height, err := strconv.Atoi(sizeSplit[0])
		if err != nil {
			return nil, fmt.Errorf("invalid height: %s", sizeSplit[0])
		}
		width, err := strconv.Atoi(sizeSplit[1])
		if err != nil {
			return nil, fmt.Errorf("invalid width: %s", sizeSplit[1])
		}
		hyperbolicReq.Height = height
		hyperbolicReq.Width = width
	} else {
		hyperbolicReq.Height = 1024
		hyperbolicReq.Width = 1024
	}

	hyperbolicReq.Backend = "auto"

	return &hyperbolicReq, nil
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
