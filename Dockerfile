FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/videra ./cmd/videra

FROM alpine:3.21 AS runtime-base

RUN apk add --no-cache ffmpeg ca-certificates

WORKDIR /app
COPY --from=builder /out/videra /usr/local/bin/videra

FROM runtime-base AS runtime-slim

ENV VIDERA_TRANSPORT=stdio
ENV VIDERA_HTTP_ADDR=:8080
ENV VIDERA_DATA_DIR=/data
ENV VIDERA_RUNTIME_MODE=local
ENV VIDERA_FRAME_INTERVAL_SEC=5
ENV VIDERA_DEFAULT_SEARCH_LIMIT=5
ENV VIDERA_INDEX_CONCURRENCY=4
ENV VIDERA_SEARCH_AUDIO_WEIGHT=1.0
ENV VIDERA_SEARCH_VISUAL_WEIGHT=1.0

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/videra"]

FROM debian:bookworm-slim AS runtime-full

RUN apt-get update \
	&& apt-get install -y --no-install-recommends \
		ca-certificates \
		ffmpeg \
		python3 \
		python3-pip \
		tesseract-ocr \
	&& rm -rf /var/lib/apt/lists/*

RUN python3 -m pip install --break-system-packages --no-cache-dir --upgrade pip \
	&& python3 -m pip install --break-system-packages --no-cache-dir openai-whisper

WORKDIR /app
COPY --from=builder /out/videra /usr/local/bin/videra

ENV VIDERA_TRANSPORT=stdio
ENV VIDERA_HTTP_ADDR=:8080
ENV VIDERA_DATA_DIR=/data
ENV VIDERA_RUNTIME_MODE=local
ENV VIDERA_FRAME_INTERVAL_SEC=5
ENV VIDERA_DEFAULT_SEARCH_LIMIT=5
ENV VIDERA_INDEX_CONCURRENCY=4
ENV VIDERA_SEARCH_AUDIO_WEIGHT=1.0
ENV VIDERA_SEARCH_VISUAL_WEIGHT=1.0

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/videra"]

FROM runtime-slim AS runtime
