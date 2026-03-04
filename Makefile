.PHONY: build test integration-test docker-build run-stdio run-http local-up local-down local-smoke

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
		echo "Usage: make local-smoke VIDEO=/videos/<file>.mp4 [QUERY='budget roadmap'] [ENDPOINT=http://localhost:8080/mcp]"; \
		exit 1; \
	fi
	@ENDPOINT_VALUE="${ENDPOINT:-http://localhost:8080/mcp}"; \
	QUERY_VALUE="${QUERY:-budget roadmap}"; \
	go run ./cmd/localsmoke --endpoint "$$ENDPOINT_VALUE" --video "$(VIDEO)" --query "$$QUERY_VALUE"
