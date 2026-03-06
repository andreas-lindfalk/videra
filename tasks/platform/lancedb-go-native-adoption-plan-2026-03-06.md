# LanceDB Go Native Adoption Plan (2026-03-06)

## Why this exists

`github.com/lancedb/lancedb-go` is available and supports local file-backed LanceDB usage. Adopting it safely in this repo requires explicit native packaging work because the SDK depends on CGO + platform-specific native artifacts.

## Preconditions

- Choose a packaging strategy for native LanceDB artifacts (`include/lancedb.h` + per-platform libs).
- Define CI strategy for macOS/Linux builds where those artifacts are required.
- Define container strategy for both runtime profiles (`runtime-slim` and `runtime-full`) if native mode is enabled.

## Suggested implementation sequence

1. Add optional native driver boundary in storage adapter (`python` bridge remains fallback path).
2. Gate native implementation behind explicit build tag (for example `lancedb_native`) to avoid breaking zero-dependency default builds.
3. Add `VIDERA_LANCEDB_DRIVER` runtime selector (`python|native`) and fail fast when `native` is requested but native build support is not present.
4. Add focused integration test proving local file-backed LanceDB indexing/search in native mode.
5. Add benchmark capture comparing `chromem`, `lancedb`-python-bridge, and `lancedb`-native where supported.

## Release gate requirement

Do not promote native mode as default until:

- `make test` and integration coverage pass with native artifacts present,
- docker runtime image(s) can build reproducibly with native artifacts,
- rollback path (`VIDERA_STORAGE_BACKEND=chromem`) is verified in the same environment.
