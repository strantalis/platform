# Shared-Library TDF Writer (c-shared)

This package exposes the experimental Go TDF writer as a native shared library so any FFI consumer can mint a TDF without booting the WASM module. Build it with Go's `-buildmode=c-shared` target to get both `libtdf.so` and the matching `libtdf.h` header.

```bash
GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) \
  go build -buildmode=c-shared -o libtdf.so ./examples/wasm/tdf_poc/cshared
# Produces libtdf.so and libtdf.h next to the command output path
```

## Exported API

`libtdf.h` mirrors the WASM module entry points so host code can drive the writer step-by-step:

| Symbol | Signature (simplified) | Description |
| --- | --- | --- |
| `int32_t TDFNewWriter(const uint8_t* kas_json, size_t kas_len)` | Parses an optional default KAS JSON blob (same schema as `policy.SimpleKasKey`) and prepares a fresh writer instance. |
| `int32_t TDFWriteSegment(int32_t index, const uint8_t* data, size_t data_len, uint8_t* out_buf, size_t out_cap)` | Encrypts a segment, writes the archive bytes into the caller-provided buffer, and returns the number of bytes produced. Also populates `TDFGetLastSegmentInfo/Data`. |
| `int32_t TDFFinalize(const uint8_t* attrs_json, size_t attrs_len, uint8_t* out_buf, size_t out_cap)` | Finalizes the archive using the supplied attribute array (same shape as `[]*policy.Value`) and writes the trailing ZIP bytes into `out_buf`. Updates `TDFGetFinalizeInfo/Manifest`. |
| `int32_t TDFGetLastSegmentInfo(uint8_t* out_buf, size_t out_cap)` | Copies the most recent segment metadata JSON into `out_buf` (or returns the required size if `out_buf == NULL`). |
| `int32_t TDFGetLastSegmentData(uint8_t* out_buf, size_t out_cap)` | Copies the raw segment bytes returned by the previous `TDFWriteSegment`. |
| `int32_t TDFGetFinalizeInfo(uint8_t* out_buf, size_t out_cap)` | Copies finalize metrics (segment counts, sizes, etc.) as JSON. |
| `int32_t TDFGetFinalizeManifest(uint8_t* out_buf, size_t out_cap)` | Copies the manifest JSON from the most recent finalize. |
| `int32_t TDFGetLastError(char* out_buf, size_t out_cap)` | Retrieves the last error message as a null-terminated UTF-8 string; pass `NULL, 0` to query the length first. |
| `int32_t TDFReset(void)` | Clears all cached state and releases the active writer (mirrors the WASM `TDFReset`). |

All output buffers follow the same convention as the WASM module:

* Passing `NULL` or `cap=0` returns the required size (in bytes) without copying data.
* If the provided buffer is too small, the call returns `statusBufferTooSmall` and keeps `lastError` set to a helpful message.
* Successful calls return the number of bytes written (always ≥ 0).

### Status codes

| Code | Meaning | Typical cause |
| --- | --- | --- |
| `0` | Success / zero bytes written | Valid call that produced no bytes (e.g., empty payload). |
| `>0` | Success with payload | Number of bytes copied into the caller’s buffer. |
| `-1` (`statusError`) | Generic failure | JSON parse issues, crypto errors, etc. Details available via `TDFGetLastError`. |
| `-2` (`statusNoWriter`) | Writer not initialized | `TDFNewWriter` hasn’t been called or `TDFReset` cleared the state. |
| `-3` (`statusBufferTooSmall`) | Output buffer too small | Re-run with a larger buffer. |

### Calling conventions

1. Call `TDFNewWriter` once per archive (optionally passing a default KAS JSON blob).
2. Feed payload slices through `TDFWriteSegment` (one or many, any order). Concatenate the returned archive bytes in your host code—the helper does not retain them.
3. Call `TDFFinalize` with the JSON attribute array. Append the returned bytes to your archive buffer to obtain a complete TDF ZIP.
4. Optionally fetch metadata via the `TDFGet*` helpers.
5. Call `TDFReset` before starting the next archive (or rely on `TDFFinalize` clearing the active writer).

## Sample harnesses

The `samples/` directory demonstrates how to call the staged API from Python (`ctypes`) and Node (`ffi-napi`). Both harnesses:

1. Load the JSON fixtures in `testdata/`.
2. Call `TDFNewWriter`.
3. Stream a single payload chunk through `TDFWriteSegment`.
4. Finalize, append the returned bytes, and write the full `.tdf`.

```
# Python (ctypes)
python3 examples/wasm/tdf_poc/cshared/samples/python/run.py --out demo-python.tdf

# Node (ffi-napi)
cd examples/wasm/tdf_poc/cshared/samples/node
npm install ffi-napi ref-napi
node run.cjs --out demo-node.tdf
```

## Benchmark context

- `examples/wasm/tdf_poc/README.md` lists the canonical host vs. module vs. native numbers.
- `examples/wasm/tdf_poc/cshared/bench` adds a Go benchmark that loads `libtdf.so` via `dlopen` and drives the exported C ABI, so you can quantify the CGO/FFI overhead directly:

  ```bash
  go test -run=^$ -bench=BenchmarkSharedLibraryFFI -benchmem -benchtime=1x ./examples/wasm/tdf_poc/cshared/bench
  ```

  (Feel free to raise `-benchtime` once you’re confident with the runtime; on slower machines the default can take a while.)

  When explaining the numbers to non-engineers, borrow the same analogy as the main README: encrypting via WASM-only is like shipping a parcel by boat, the shared library is the express train, and the native Go path is the lab stopwatch. All paths obey the same security rules; we’re simply swapping the transport layer to get results faster for Python/Node callers.

  ```mermaid
  %%{init: {theme: neutral, logLevel: fatal}}%%
  chart LR
      title Average encryption time (ms)
      x-axis Path
      y-axis Time (ms)
      "Shared native" : 0.037
      "Shared FFI" : 0.066
      "WASM host" : 0.94
      "WASM module" : 1.64
  ```
