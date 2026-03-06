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

FROM golang:1.25-bookworm AS builder-lancedb-native

WORKDIR /app

ARG LANCEDB_NATIVE_VERSION=v0.1.2
ARG TARGETARCH

RUN apt-get update \
	&& apt-get install -y --no-install-recommends \
		ca-certificates \
		curl \
		bash \
		build-essential \
		tar \
	&& rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN curl -fsSL -o /tmp/lancedb-go-native-binaries.tar.gz "https://github.com/lancedb/lancedb-go/releases/download/${LANCEDB_NATIVE_VERSION}/lancedb-go-native-binaries.tar.gz" \
	&& tar -xzf /tmp/lancedb-go-native-binaries.tar.gz -C /app \
	&& rm -f /tmp/lancedb-go-native-binaries.tar.gz

RUN ARCH="${TARGETARCH}"; \
	if [ -z "$ARCH" ]; then ARCH="$(dpkg --print-architecture)"; fi; \
	case "$ARCH" in \
		amd64|arm64) ;; \
		aarch64) ARCH="arm64" ;; \
		x86_64) ARCH="amd64" ;; \
		*) echo "unsupported TARGETARCH: $ARCH"; exit 1 ;; \
	esac; \
	if [ ! -f "/app/lib/linux_${ARCH}/liblancedb_go.a" ]; then \
		echo "missing LanceDB native artifact for linux_${ARCH}; use --platform linux/amd64 with LANCEDB_NATIVE_VERSION=${LANCEDB_NATIVE_VERSION}"; \
		exit 1; \
	fi; \
	ranlib "/app/lib/linux_${ARCH}/liblancedb_go.a"; \
	CGO_ENABLED=1 GOOS=linux GOARCH="$ARCH" \
	CGO_CFLAGS="-I/app/include" \
	CGO_LDFLAGS="/app/lib/linux_${ARCH}/liblancedb_go.a -lm -ldl -lpthread" \
	go build -tags lancedb_native -o /out/videra ./cmd/videra

FROM debian:bookworm-slim AS runtime-lancedb-native

RUN apt-get update \
	&& apt-get install -y --no-install-recommends \
		ca-certificates \
		ffmpeg \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder-lancedb-native /out/videra /usr/local/bin/videra

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
