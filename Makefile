.PHONY: build test integration-test docker-build run-stdio run-http local-up local-down local-smoke local-smoke-default local-e2e

build:
	go build -o bin/videra ./cmd/videra

test:
	go test ./...

integration-test:
	go test ./test/integration/... -v -tags=integration -timeout=180s

docker-build:
	docker build -t videra:dev .

run-stdio: docker-build
	docker run -i --rm videra:dev

run-http: docker-build
	docker run --rm -p 8080:8080 -e VIDERA_TRANSPORT=http videra:dev

local-up:
	docker compose up -d --build

local-down:
	docker compose down

local-smoke:
	@if [ -z "$(VIDEO)" ]; then \
		echo "Usage: make local-smoke VIDEO=/videos/<file>.mp4 [QUERY='test query'] [ENDPOINT=http://localhost:8080/mcp]"; \
		exit 1; \
	fi
	@go run ./cmd/localsmoke \
		--endpoint "$(or $(ENDPOINT),http://localhost:8080/mcp)" \
		--video "$(VIDEO)" \
		--query "$(or $(QUERY),test query)"

local-smoke-default:
	@if [ ! -d "./videos" ]; then \
		echo "Missing ./videos folder. Create it and place a video file inside."; \
		exit 1; \
	fi
	@VIDEO_FILE="$$(find ./videos -maxdepth 1 -type f \( -iname '*.mov' -o -iname '*.mp4' -o -iname '*.m4v' -o -iname '*.mkv' \) | head -n 1)"; \
	if [ -z "$$VIDEO_FILE" ]; then \
		echo "No local video found in ./videos (supported: .mov, .mp4, .m4v, .mkv)"; \
		echo "Place a file in ./videos and run make local-smoke-default again."; \
		exit 1; \
	fi; \
	BASENAME="$$(basename "$$VIDEO_FILE")"; \
	echo "Using ./videos/$$BASENAME"; \
	$(MAKE) local-smoke VIDEO="/videos/videos/$$BASENAME" QUERY="$(or $(QUERY),test query)" ENDPOINT="$(or $(ENDPOINT),http://localhost:8080/mcp)"

local-e2e: local-up local-smoke-default
	@echo "Local MCP server is running at http://localhost:8080/mcp"
	@echo "Stop it with: make local-down"
