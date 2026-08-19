# Build the Go service with all rendering assets embedded in the binary.
FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/typlyd ./cmd/typlyd

# ffmpeg is required only for MP4 exports. GIF exports use the Go binary.
FROM alpine:3.22

RUN apk add --no-cache ca-certificates ffmpeg
COPY --from=build /out/typlyd /usr/local/bin/typlyd

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/typlyd"]
