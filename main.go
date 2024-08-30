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

    hyperbolicRequest.ModelName = openAIRequest.Model
    hyperbolicRequest.Prompt = openAIRequest.Prompt

    var sizeSplit = strings.Split(openAIRequest.Size, "x")
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

    w.Write(body)
}

func main() {
    router := mux.NewRouter()
    router.HandleFunc("/image/generation", imageGenerationHandler).Methods("POST")

    fmt.Println("Server is running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", router))
}
