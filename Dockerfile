FROM --platform=$BUILDPLATFORM golang:1.22 as build
WORKDIR /go/src/app
ADD . /go/src/app

RUN go get -d -v ./...
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/app github.com/cheahjs/hyperbolic-openai-proxy

FROM --platform=$TARGETPLATFORM gcr.io/distroless/static-debian11:debug
COPY --from=build /go/bin/app /
ENTRYPOINT ["/app"]
