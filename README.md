# hyperbolic-openai-proxy
Call [Hyperbolic](https://app.hyperbolic.xyz) image generation endpoints with OpenAI-compatible endpoints

## Deployment

### Using `docker run`

To deploy using `docker run`, you can use the following command:

```sh
docker run -d -p 8080:8080 ghcr.io/cheahjs/hyperbolic-openai-proxy:latest
```

### Using `docker compose`

To deploy using `docker compose`, you can use the provided `docker-compose.yaml` file. First, create a `.env` file with the following content:

```sh
LISTEN_ADDR=:8080
```

Then, run the following command:

```sh
docker-compose up -d
```

### Using `go install`

To deploy using `go install`, you need to have Go installed on your machine. Run the following commands:

```sh
go install github.com/cheahjs/hyperbolic-openai-proxy@latest
LISTEN_ADDR=:8080 hyperbolic-openai-proxy
```
