---
status: accepted
date: '2025-11-14'
tags:
 - sdk
 - wasm
 - ffi
driver: strantalis
deciders: platform-sdk
---
# Expose the experimental TDF writer through a c-shared library that mirrors the WASM API

## Context and Problem Statement

The WASM proof-of-concept already exposes the experimental writer via `TDFNewWriter/TDFWriteSegment/TDFFinalize`, but Python and Node consumers pay a heavy runtime penalty by instantiating Wasmtime and proxying crypto calls back to the host. We need a native shared-library artifact that any FFI client can load without re-implementing the writer logic or diverging from the existing WASM contract.

## Decision Drivers

* Minimize end-to-end latency for Python/Node clients by avoiding the WASM runtime.
* Keep the host-facing API identical to the module exports so existing bindings and tooling stay consistent.
* Provide hard performance data comparing the c-shared path with the WASM host/module flows.

## Considered Options

* Continue using the WASM module + hostcrypto bridge.
* Embed the Go writer in each language runtime separately (Python extension, Node addon, etc.).
* Build a `-buildmode=c-shared` library that mirrors the WASM exports and reuse it across FFIs. ✅

## Decision Outcome

Chosen option: **c-shared library that mirrors the WASM API**, because it gives us near-native performance, keeps the existing entry points, and can be used directly from Python/Node (or any FFI) without bespoke language-specific ports.

### Consequences

* 🟩 **Good** — Python/Node harnesses can stream segments and finalize TDFs with the same calls they already know from the WASM runner.
* 🟩 **Good** — Benchmarks show the shared library path is ~12× faster than the WASM host bridge and ~24× faster than the pure module path.
* 🟥 **Bad** — CGO boundaries require explicit buffer management (callers must retry on `statusBufferTooSmall`) and we need to ship per-platform binaries.

## Validation

* `python3 examples/wasm/tdf_poc/cshared/samples/python/run.py --lib examples/wasm/tdf_poc/cshared/libtdf.so --out /tmp/python-tdf.tdf`
* `unzip -l /tmp/python-tdf.tdf`
* `GOCACHE=$(pwd)/.gocache go test -bench=BenchmarkCrypto -run=^$ ./examples/wasm/tdf_poc/host`
* `GOCACHE=$(pwd)/.gocache go test -run=^$ -bench=BenchmarkSharedLibraryFFI -benchmem -benchtime=1x ./examples/wasm/tdf_poc/cshared/bench`

## Pros and Cons of the Options

### c-shared library mirroring WASM (selected)

* 🟩 **Good**, because the API parity makes it trivial to port existing bindings.
* 🟩 **Good**, because the CGO boundary is far cheaper than spinning up Wasmtime + hostcrypto for every run (≈66 µs/segment/finalize vs. 0.94 ms for host-WASM and 1.64 ms for module-WASM).
* 🟨 **Neutral**, because we must ship platform-specific binaries (`libtdf.so`, `libtdf.dylib`, `tdf.dll`), but that is manageable through CI.
* 🟥 **Bad**, because CGO rules prevent us from returning Go-owned buffers, so FFI callers must manage retry loops and output buffers explicitly.

### Status quo: WASM module + hostcrypto bridge

* 🟩 **Good**, because it is portable and already implemented.
* 🟥 **Bad**, because every run instantiates Wasmtime and proxies crypto back to Go, resulting in ~0.94 ms (host mode) to ~1.64 ms (module mode) per run and extra complexity in Python/Node hosts.

### Per-language native embeds

* 🟩 **Good**, because each language could optimize for its ABI.
* 🟥 **Bad**, because it duplicates writer logic, risks divergence, and adds a maintenance burden for every runtime we support.

## More Information

Benchmark results gathered on 2025-11-14 (macOS 15.1, Apple M3 Max, Go 1.24.9):

| Path | Command | ns/op | B/op | allocs/op |
| --- | --- | --- | --- | --- |
| c-shared FFI (`dlopen` + exported API) | `go test -run=^$ -bench=BenchmarkSharedLibraryFFI -benchmem -benchtime=1x ./examples/wasm/tdf_poc/cshared/bench` | **66,375** | 131,072 | 2 |
| Native Go control | `go test -bench=BenchmarkCrypto -run=^$ ./examples/wasm/tdf_poc/host` → `BenchmarkCryptoShared` | 37,060 | 24,489 | 179 |
| WASM + host crypto bridge | same command → `BenchmarkCryptoHost` | 943,589 | 23,153 | 470 |
| WASM + module crypto | same command → `BenchmarkCryptoModule` | 1,635,548 | 14,088 | 308 |

The shared library runs retain the streaming semantics (callers append `TDFWriteSegment` output + `TDFFinalize` output) and keep metadata accessible through `TDFGetLastSegmentInfo/Data` and `TDFGetFinalizeInfo/Manifest`, so future bindings can reuse the exact same integration points as the WASM runner.
