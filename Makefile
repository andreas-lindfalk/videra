.PHONY: build build-lancedb-native test integration-test integration-test-fresh integration-test-lancedb-native docker-build docker-build-slim docker-build-full docker-build-lancedb-native docker-build-lancedb-native-clip run-stdio run-http run-stdio-full run-http-full run-stdio-lancedb-native run-http-lancedb-native run-stdio-lancedb-native-clip run-http-lancedb-native-clip local-up local-down local-smoke local-smoke-default local-index-folder local-e2e release-gate release-gate-split release-gate-preflight release-gate-clean pilot-quality-gate real-corpus-promotion-gate deployment-promotion-gate storage-benchmark-gate storage-benchmark-capture storage-benchmark-summarize rollback-rehearsal-capture rollback-rehearsal-summarize gate-parity-capture gate-parity-summarize phase32-candidate-proof-pack phase32-candidate-proof-pack-summarize

build:
	go build -o bin/videra ./cmd/videra

build-lancedb-native:
	@LANCEDB_VERSION="$${LANCEDB_NATIVE_VERSION:-v0.1.2}"; \
	OS="$$(go env GOOS)"; \
	ARCH="$$(go env GOARCH)"; \
	if [ ! -f "include/lancedb.h" ] || [ ! -f "lib/$${OS}_$$ARCH/liblancedb_go.a" ]; then \
		echo "Downloading LanceDB native artifacts ($$LANCEDB_VERSION)..."; \
		curl -sSL https://raw.githubusercontent.com/lancedb/lancedb-go/main/scripts/download-artifacts.sh | bash -s -- "$$LANCEDB_VERSION"; \
	fi; \
	CGO_CFLAGS="-I$(PWD)/include" \
	CGO_LDFLAGS="$(PWD)/lib/$${OS}_$$ARCH/liblancedb_go.a $$( [ "$$OS" = "darwin" ] && echo "-framework Security -framework CoreFoundation" || echo "-lm -ldl -lpthread" )" \
	CGO_ENABLED=1 \
	go build -tags lancedb_native -o bin/videra-native ./cmd/videra

test:
	go test ./...

integration-test:
	go test ./test/integration/... -v -tags=integration -timeout=180s

integration-test-fresh:
	go test ./test/integration/... -v -tags=integration -timeout=180s -count=1

integration-test-lancedb-native:
	go test ./test/integration/... -v -tags=integration -run 'TestLanceDBNativeBackendIndexesAndSearches|TestLanceDBBackendOnDefaultRuntimeReturnsGuidanceError' -timeout=240s -count=1

docker-build: docker-build-lancedb-native

docker-build-slim:
	docker build --target runtime-slim -t videra:dev -t videra:dev-slim .

docker-build-full:
	docker build --target runtime-full -t videra:dev-full .

docker-build-lancedb-native:
	docker build --platform $${LANCEDB_DOCKER_PLATFORM:-linux/amd64} --target runtime-lancedb-native --build-arg LANCEDB_NATIVE_VERSION=$${LANCEDB_NATIVE_VERSION:-v0.1.2} -t videra:dev-lancedb-native .

docker-build-lancedb-native-clip:
	docker build --platform $${LANCEDB_DOCKER_PLATFORM:-linux/amd64} --target runtime-lancedb-native-clip --build-arg LANCEDB_NATIVE_VERSION=$${LANCEDB_NATIVE_VERSION:-v0.1.2} -t videra:dev-lancedb-native-clip .

run-stdio: docker-build-lancedb-native
	docker run -i --rm -e VIDERA_STORAGE_BACKEND=lancedb videra:dev-lancedb-native

run-http: docker-build-lancedb-native
	docker run --rm -p 8080:8080 -e VIDERA_TRANSPORT=http -e VIDERA_STORAGE_BACKEND=lancedb videra:dev-lancedb-native

run-stdio-full: docker-build-full
	docker run -i --rm videra:dev-full

run-http-full: docker-build-full
	docker run --rm -p 8080:8080 -e VIDERA_TRANSPORT=http videra:dev-full

run-stdio-lancedb-native: docker-build-lancedb-native
	docker run -i --rm -e VIDERA_STORAGE_BACKEND=lancedb videra:dev-lancedb-native

run-http-lancedb-native: docker-build-lancedb-native
	docker run --rm -p 8080:8080 -e VIDERA_TRANSPORT=http -e VIDERA_STORAGE_BACKEND=lancedb videra:dev-lancedb-native

run-stdio-lancedb-native-clip: docker-build-lancedb-native-clip
	docker run -i --rm -e VIDERA_STORAGE_BACKEND=lancedb videra:dev-lancedb-native-clip

run-http-lancedb-native-clip: docker-build-lancedb-native-clip
	docker run --rm -p 8080:8080 -e VIDERA_TRANSPORT=http -e VIDERA_STORAGE_BACKEND=lancedb videra:dev-lancedb-native-clip

local-up:
	VIDERA_DOCKER_TARGET=$${VIDERA_DOCKER_TARGET:-runtime-lancedb-native} \
	VIDERA_DOCKER_PLATFORM=$${VIDERA_DOCKER_PLATFORM:-linux/amd64} \
	VIDERA_STORAGE_BACKEND=$${VIDERA_STORAGE_BACKEND:-lancedb} \
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
	$(MAKE) local-smoke VIDEO="/videos/$$BASENAME" QUERY="$(or $(QUERY),test query)" ENDPOINT="$(or $(ENDPOINT),http://localhost:8080/mcp)"

local-index-folder:
	@HOST_VIDEO_DIR="$(or $(VIDEO_DIR),$(VIDERA_VIDEO_DIR),./videos)"; \
	if [ ! -d "$$HOST_VIDEO_DIR" ]; then \
		echo "Missing video folder: $$HOST_VIDEO_DIR"; \
		echo "Set VIDEO_DIR=/path/to/videos (or VIDERA_VIDEO_DIR) and try again."; \
		exit 1; \
	fi; \
	VIDEO_LIST="$$(find "$$HOST_VIDEO_DIR" -type f \( -iname '*.mov' -o -iname '*.mp4' -o -iname '*.m4v' -o -iname '*.mkv' \) | sort)"; \
	if [ -z "$$VIDEO_LIST" ]; then \
		echo "No videos found in $$HOST_VIDEO_DIR (supported: .mov, .mp4, .m4v, .mkv)"; \
		exit 1; \
	fi; \
	COUNT="$$(printf '%s\n' "$$VIDEO_LIST" | grep -c .)"; \
	printf '%s\n' "$$VIDEO_LIST" | while IFS= read -r VIDEO_FILE; do \
		[ -z "$$VIDEO_FILE" ] && continue; \
		REL_PATH="$${VIDEO_FILE#$$HOST_VIDEO_DIR/}"; \
		if [ "$$REL_PATH" = "$$VIDEO_FILE" ]; then REL_PATH="$$(basename "$$VIDEO_FILE")"; fi; \
		echo "Indexing $$REL_PATH"; \
		go run ./cmd/localindex --endpoint "$(or $(ENDPOINT),http://localhost:8080/mcp)" --video "/videos/$$REL_PATH"; \
	done; \
	echo "Indexed $$COUNT video(s). You can now query via MCP at $(or $(ENDPOINT),http://localhost:8080/mcp)."

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

real-corpus-promotion-gate:
	$(MAKE) pilot-quality-gate
	go test ./test/integration/... -v -tags=integration -run 'TestDefaultIntegrationSuite/(TestProofPackScenariosEvidenceAndDeterminism|TestProofPackProductRecallPrioritizesTop2Evidence|TestSearchVideoDeterministicOrdering|TestToolResponseBackwardCompatFields)' -count=1

deployment-promotion-gate:
	$(MAKE) release-gate
	$(MAKE) release-gate-split
	$(MAKE) real-corpus-promotion-gate

storage-benchmark-gate:
	go test ./internal/storage -run '^$$' -bench 'BenchmarkChromemStoreBaseline' -benchmem -count=1

storage-benchmark-capture:
	@OUT_PATH="$(or $(OUT),/tmp/videra_storage_benchmark_capture.out)"; \
	EXIT_PATH="$(or $(EXIT_OUT),/tmp/videra_storage_benchmark_capture.exit)"; \
	BACKEND_VALUE="$(or $(BACKEND),chromem)"; \
	date > "$$OUT_PATH"; \
	echo "benchmark_backend=$$BACKEND_VALUE" >> "$$OUT_PATH"; \
	VIDERA_STORAGE_BACKEND="$$BACKEND_VALUE" $(MAKE) storage-benchmark-gate >> "$$OUT_PATH" 2>&1; \
	echo $$? > "$$EXIT_PATH"; \
	echo "Benchmark output: $$OUT_PATH"; \
	echo "Exit code file: $$EXIT_PATH"

storage-benchmark-summarize:
	@OUT_PATH="$(or $(OUT),/tmp/videra_storage_benchmark_capture.out)"; \
	grep -E 'BenchmarkChromemStoreBaseline|ns/op|B/op|allocs/op|PASS|FAIL' "$$OUT_PATH"

rollback-rehearsal-capture:
	@OUT_PATH="$(or $(OUT),/tmp/videra_phase32_prereq4_rollback_rehearsal.out)"; \
	EXIT_PATH="$(or $(EXIT_OUT),/tmp/videra_phase32_prereq4_rollback_rehearsal.exit)"; \
	BACKEND_VALUE="$(or $(BACKEND),chromem)"; \
	date > "$$OUT_PATH"; \
	echo "rollback_rehearsal_backend=$$BACKEND_VALUE" >> "$$OUT_PATH"; \
	echo "== Step 1: Pre-rollback gate (release-gate-split) ==" >> "$$OUT_PATH"; \
	VIDERA_STORAGE_BACKEND="$$BACKEND_VALUE" $(MAKE) release-gate-split >> "$$OUT_PATH" 2>&1; \
	PRE_RC=$$?; \
	echo "pre_release_gate_split_exit=$$PRE_RC" >> "$$OUT_PATH"; \
	if [ $$PRE_RC -ne 0 ]; then \
		echo $$PRE_RC > "$$EXIT_PATH"; \
		echo "Rollback rehearsal failed at pre-rollback gate."; \
		exit $$PRE_RC; \
	fi; \
	echo "== Step 2: Simulated rollback to stable backend and rerun gate ==" >> "$$OUT_PATH"; \
	VIDERA_STORAGE_BACKEND=chromem $(MAKE) release-gate-split >> "$$OUT_PATH" 2>&1; \
	ROLLBACK_RC=$$?; \
	echo "post_rollback_release_gate_split_exit=$$ROLLBACK_RC" >> "$$OUT_PATH"; \
	echo $$ROLLBACK_RC > "$$EXIT_PATH"; \
	if [ $$ROLLBACK_RC -eq 0 ]; then \
		echo "Rollback rehearsal output: $$OUT_PATH"; \
		echo "Exit code file: $$EXIT_PATH"; \
	else \
		echo "Rollback rehearsal failed at post-rollback gate."; \
		exit $$ROLLBACK_RC; \
	fi

rollback-rehearsal-summarize:
	@OUT_PATH="$(or $(OUT),/tmp/videra_phase32_prereq4_rollback_rehearsal.out)"; \
	grep -E 'Step|PASS|FAIL|ok\s+github.com/andreas-lindfalk/videra/test/integration|pre_release_gate_split_exit|post_rollback_release_gate_split_exit' "$$OUT_PATH"

gate-parity-capture:
	@OUT_PATH="$(or $(OUT),/tmp/videra_phase32_prereq5_gate_parity.out)"; \
	EXIT_PATH="$(or $(EXIT_OUT),/tmp/videra_phase32_prereq5_gate_parity.exit)"; \
	BACKEND_VALUE="$(or $(BACKEND),chromem)"; \
	OVERALL_RC=0; \
	date > "$$OUT_PATH"; \
	echo "== Step 1: release-gate ==" >> "$$OUT_PATH"; \
	VIDERA_STORAGE_BACKEND="$$BACKEND_VALUE" $(MAKE) release-gate >> "$$OUT_PATH" 2>&1; \
	RC=$$?; \
	echo "release_gate_exit=$$RC" >> "$$OUT_PATH"; \
	if [ $$RC -ne 0 ]; then OVERALL_RC=$$RC; fi; \
	echo "== Step 2: release-gate-split ==" >> "$$OUT_PATH"; \
	VIDERA_STORAGE_BACKEND="$$BACKEND_VALUE" $(MAKE) release-gate-split >> "$$OUT_PATH" 2>&1; \
	RC=$$?; \
	echo "release_gate_split_exit=$$RC" >> "$$OUT_PATH"; \
	if [ $$RC -ne 0 ] && [ $$OVERALL_RC -eq 0 ]; then OVERALL_RC=$$RC; fi; \
	echo "== Step 3: pilot-quality-gate ==" >> "$$OUT_PATH"; \
	VIDERA_STORAGE_BACKEND="$$BACKEND_VALUE" $(MAKE) pilot-quality-gate >> "$$OUT_PATH" 2>&1; \
	RC=$$?; \
	echo "pilot_quality_gate_exit=$$RC" >> "$$OUT_PATH"; \
	if [ $$RC -ne 0 ] && [ $$OVERALL_RC -eq 0 ]; then OVERALL_RC=$$RC; fi; \
	echo "== Step 4: real-corpus-promotion-gate ==" >> "$$OUT_PATH"; \
	VIDERA_STORAGE_BACKEND="$$BACKEND_VALUE" $(MAKE) real-corpus-promotion-gate >> "$$OUT_PATH" 2>&1; \
	RC=$$?; \
	echo "real_corpus_promotion_gate_exit=$$RC" >> "$$OUT_PATH"; \
	if [ $$RC -ne 0 ] && [ $$OVERALL_RC -eq 0 ]; then OVERALL_RC=$$RC; fi; \
	echo $$OVERALL_RC > "$$EXIT_PATH"; \
	if [ $$OVERALL_RC -eq 0 ]; then \
		echo "Gate parity output: $$OUT_PATH"; \
		echo "Exit code file: $$EXIT_PATH"; \
	else \
		echo "Gate parity capture finished with failures. See $$OUT_PATH"; \
		exit $$OVERALL_RC; \
	fi

gate-parity-summarize:
	@OUT_PATH="$(or $(OUT),/tmp/videra_phase32_prereq5_gate_parity.out)"; \
	grep -E 'Step|release_gate_exit|release_gate_split_exit|pilot_quality_gate_exit|real_corpus_promotion_gate_exit|pilot benchmark scorecard|PASS|FAIL|ok\s+github.com/andreas-lindfalk/videra/test/integration' "$$OUT_PATH"

phase32-candidate-proof-pack:
	@BACKEND_VALUE="$(BACKEND)"; \
	set -e; \
	if [ -z "$$BACKEND_VALUE" ]; then \
		echo "Usage: make phase32-candidate-proof-pack BACKEND=<candidate-backend> [PREFIX=/tmp/videra_phase32_candidate]"; \
		exit 1; \
	fi; \
	if [ "$$BACKEND_VALUE" = "chromem" ]; then \
		echo "BACKEND=chromem is baseline mode. Provide a non-baseline candidate backend."; \
		exit 1; \
	fi; \
	PREFIX_VALUE="$(or $(PREFIX),/tmp/videra_phase32_candidate)"; \
	$(MAKE) storage-benchmark-capture BACKEND="$$BACKEND_VALUE" OUT="$$PREFIX_VALUE"_prereq1_benchmark.out EXIT_OUT="$$PREFIX_VALUE"_prereq1_benchmark.exit; \
	$(MAKE) rollback-rehearsal-capture BACKEND="$$BACKEND_VALUE" OUT="$$PREFIX_VALUE"_prereq4_rollback.out EXIT_OUT="$$PREFIX_VALUE"_prereq4_rollback.exit; \
	$(MAKE) gate-parity-capture BACKEND="$$BACKEND_VALUE" OUT="$$PREFIX_VALUE"_prereq5_gate_parity.out EXIT_OUT="$$PREFIX_VALUE"_prereq5_gate_parity.exit; \
	echo "Phase 32 candidate proof pack captured with prefix: $$PREFIX_VALUE"

phase32-candidate-proof-pack-summarize:
	@PREFIX_VALUE="$(or $(PREFIX),/tmp/videra_phase32_candidate)"; \
	echo "== Prerequisite 1: benchmark summary =="; \
	$(MAKE) storage-benchmark-summarize OUT="$$PREFIX_VALUE"_prereq1_benchmark.out; \
	echo "== Prerequisite 4: rollback rehearsal summary =="; \
	$(MAKE) rollback-rehearsal-summarize OUT="$$PREFIX_VALUE"_prereq4_rollback.out; \
	echo "== Prerequisite 5: gate parity summary =="; \
	$(MAKE) gate-parity-summarize OUT="$$PREFIX_VALUE"_prereq5_gate_parity.out
