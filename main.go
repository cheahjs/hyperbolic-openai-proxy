package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"

    "github.com/gorilla/mux"
)

// OpenAIRequest represents an OpenAI image generation request
type OpenAIRequest struct {
    Model  string `json:"model"`
    Prompt string `json:"prompt"`
    N      int    `json:"n"`
    Size   string `json:"size"`
}

// HyperbolicRequest represents a Hyperbolic image generation request
type HyperbolicRequest struct {
    ModelName string `json:"model_name"`
    Prompt    string `json:"prompt"`
    Height    int    `json:"height"`
    Width     int    `json:"width"`
    Backend   string `json:"backend"`
}

func convertRequest(openAIRequest OpenAIRequest) HyperbolicRequest {
    var hyperbolicRequest HyperbolicRequest

    switch openAIRequest.Model {
    case "dall-e-2":
        hyperbolicRequest.ModelName = "SD2"
    case "dall-e-3":
        hyperbolicRequest.ModelName = "SDXL1.0-base"
    default:
        log.Println("Unsupported model")
        return hyperbolicRequest
    }

    hyperbolicRequest.Prompt = openAIRequest.Prompt

    switch openAIRequest.Size {
    case "256x256":
        hyperbolicRequest.Height = 256
        hyperbolicRequest.Width = 256
    case "512x512":
        hyperbolicRequest.Height = 512
        hyperbolicRequest.Width = 512
    case "1024x1024":
        hyperbolicRequest.Height = 1024
        hyperbolicRequest.Width = 1024
    case "1792x1024":
        hyperbolicRequest.Height = 1792
        hyperbolicRequest.Width = 1024
    case "1024x1792":
        hyperbolicRequest.Height = 1024
        hyperbolicRequest.Width = 1792
    default:
        log.Println("Unsupported size")
        return hyperbolicRequest
    }

    hyperbolicRequest.Backend = "auto"

    return hyperbolicRequest
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

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer YOUR_HYPERBOLIC_API_KEY")

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

    w.Write(body)
}

func main() {
    router := mux.NewRouter()
    router.HandleFunc("/image/generation", imageGenerationHandler).Methods("POST")

    fmt.Println("Server is running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", router))
}
