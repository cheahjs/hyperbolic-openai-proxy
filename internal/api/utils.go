package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

)

type OpenAIImage struct {
	URL     string `json:"url,omitempty"`
	B64JSON string `json:"b64_json,omitempty"`
}

type OpenAIResponse struct {
	Created int64         `json:"created"`
	Data    []OpenAIImage `json:"data"`
}

func convertRequest(req *OpenAIRequest) (*HyperbolicRequest, error) {
	var hyperbolicReq HyperbolicRequest

	hyperbolicReq.ModelName = req.Model
	hyperbolicReq.Prompt = req.Prompt

	if req.Size != nil {
		sizeSplit := strings.Split(*req.Size, "x")
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

func convertResponse(hyperbolicResponse HyperbolicResponse, openAIRequest OpenAIRequest, baseURL string, imageManager *ImageManager) (OpenAIResponse, error) {
	var openAIResponse OpenAIResponse

	openAIResponse.Created = time.Now().Unix()

	for _, image := range hyperbolicResponse.Images {
		var openAIImage OpenAIImage

		filePath, err := imageManager.StoreImageWithPrompt(openAIRequest.Prompt, []byte(image.Image))
		if err != nil {
			return openAIResponse, fmt.Errorf("failed to store image: %w", err)
		}
		openAIImage.URL = filePath

		if openAIRequest.ResponseFormat != nil && *openAIRequest.ResponseFormat == "b64_json" {
			openAIImage.B64JSON = image.Image
		}
		openAIResponse.Data = append(openAIResponse.Data, openAIImage)
	}

	return openAIResponse, nil
}
