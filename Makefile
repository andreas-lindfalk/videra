.PHONY: build test integration-test integration-test-fresh docker-build docker-build-slim docker-build-full run-stdio run-http run-stdio-full run-http-full local-up local-down local-smoke local-smoke-default local-e2e release-gate release-gate-split release-gate-preflight release-gate-clean pilot-quality-gate

build:
	go build -o bin/videra ./cmd/videra

test:
	go test ./...

integration-test:
	go test ./test/integration/... -v -tags=integration -timeout=180s

integration-test-fresh:
	go test ./test/integration/... -v -tags=integration -timeout=180s -count=1

docker-build: docker-build-slim

docker-build-slim:
	docker build --target runtime-slim -t videra:dev -t videra:dev-slim .

docker-build-full:
	docker build --target runtime-full -t videra:dev-full .

run-stdio: docker-build-slim
	docker run -i --rm videra:dev

run-http: docker-build-slim
	docker run --rm -p 8080:8080 -e VIDERA_TRANSPORT=http videra:dev

run-stdio-full: docker-build-full
	docker run -i --rm videra:dev-full

run-http-full: docker-build-full
	docker run --rm -p 8080:8080 -e VIDERA_TRANSPORT=http videra:dev-full

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

release-gate:
	$(MAKE) release-gate-preflight
	$(MAKE) build
	$(MAKE) test
	$(MAKE) integration-test-fresh
	$(MAKE) docker-build

release-gate-split:
	go test ./test/integration/... -v -tags=integration -run 'TestIndexVideoAsyncSplitRoleRedisLifecycle|TestIndexVideoAsyncSplitRoleRedisSharedStorageVisibility|TestWorkerRoleWithHTTPTransportFailsFastAtStartup' -count=1

release-gate-preflight:
	@echo "== Release Gate Preflight =="
	@echo "Workspace disk:"
	@df -h .
	@echo "Docker disk summary:"
	@docker system df

release-gate-clean:
	@echo "== Release Gate Cleanup (safe prune) =="
	@docker builder prune -f
	@docker image prune -f
	@echo "Cleanup complete. Re-run with: make release-gate && make release-gate-split"

pilot-quality-gate:
	go test ./test/integration/... -v -tags=integration -run 'TestDefaultIntegrationSuite/TestPilotBenchmarkScorecard|TestIndexVideoRealMode(RemotePathRespectsMaxSizeBound|RemotePathHonorsDisabledFetch|RequiresSidecarForLocalPath)' -count=1
