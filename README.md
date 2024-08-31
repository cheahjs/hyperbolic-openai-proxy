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
BASE_URL= # Optional: Set the base URL for the proxy, e.g., https://my-proxy.com
BASE_URL_SCHEME= # Optional: Set the scheme for the base URL (http or https), defaults to http if not specified
IMAGES_SAVE_PATH= # Optional: Set the path to save generated images to the filesystem. If not set, images are stored in-memory.
IMAGE_EXPIRY=30m # Optional: Set the expiry time for cached images, defaults to 30 minutes.
MAX_IMAGE_STORE_SIZE_MB=50 # Optional: Set the maximum size of the in-memory image store in MB, defaults to 50 MB.
LOG_LEVEL=info # Optional: Set the log level (debug, info, warn, error, fatal, panic), defaults to info.
```

Then, run the following command:

```sh
docker-compose up -d
```

### Using `go install`

To deploy using `go install`, you need to have Go installed on your machine. Run the following commands:

```sh
go install github.com/cheahjs/hyperbolic-openai-proxy@latest
LISTEN_ADDR=:8080 BASE_URL=https://my-proxy.com IMAGES_SAVE_PATH=/tmp/images hyperbolic-openai-proxy
```
