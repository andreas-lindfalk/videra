.PHONY: build test integration-test docker-build run-stdio run-http

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
