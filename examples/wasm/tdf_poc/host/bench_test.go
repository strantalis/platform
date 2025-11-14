package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/opentdf/platform/protocol/go/policy"
	"github.com/opentdf/platform/sdk/experimental/tdf"
)

func buildWasmArtifact(b *testing.B) string {
	b.Helper()

	tempDir := b.TempDir()
	wasmOut := filepath.Join(tempDir, "writer.wasm")
	cacheDir := filepath.Join(tempDir, "gocache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		b.Fatalf("create cache dir: %v", err)
	}

	cmd := exec.Command("go", "build", "-o", wasmOut, "../module")
	cmd.Env = append(os.Environ(),
		"GOOS=wasip1",
		"GOARCH=wasm",
		"GOCACHE="+cacheDir,
	)
	cmd.Dir = "."

	if output, err := cmd.CombinedOutput(); err != nil {
		b.Fatalf("build wasm module: %v\n%s", err, output)
	}

	return wasmOut
}

func benchmarkCrypto(b *testing.B, mode int32) {
	wasmPath := buildWasmArtifact(b)

	rt, cleanup, err := loadWasmRuntime(wasmPath)
	if err != nil {
		b.Fatalf("load runtime: %v", err)
	}
	defer cleanup()

	kasKey := demoKasKey()
	attrs := demoAttributes(kasKey)
	payload := []byte("hello wasm tdf")

	kasJSON := mustJSON(kasKey)
	attrsJSON := mustJSON(attrs)

	rt.setMode(mode)
	if reported := rt.mode(); reported != mode {
		b.Fatalf("unexpected crypto mode reported: got %d want %d", reported, mode)
	}

	origLog := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(origLog)

	rt.setLogging(false)
	// Warm-up
	rt.run(kasJSON, attrsJSON, payload, int32(256*1024), false)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rt.run(kasJSON, attrsJSON, payload, int32(256*1024), false)
	}
}

func BenchmarkCryptoHost(b *testing.B) {
	benchmarkCrypto(b, cryptoModeHost)
}

func BenchmarkCryptoModule(b *testing.B) {
	benchmarkCrypto(b, cryptoModeModule)
}

var sharedSink int64

func BenchmarkCryptoShared(b *testing.B) {
	ctx := context.Background()
	kasKey := demoKasKey()
	attrs := demoAttributes(kasKey)
	payload := []byte("hello wasm tdf")

	var sink int64

	// warm-up
	if err := nativeRunOnce(ctx, kasKey, attrs, payload, &sink); err != nil {
		b.Fatalf("warmup failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := nativeRunOnce(ctx, kasKey, attrs, payload, &sink); err != nil {
			b.Fatalf("shared run failed: %v", err)
		}
	}

	sharedSink = sink
}

func nativeRunOnce(ctx context.Context, kasKey *policy.SimpleKasKey, attrs []*policy.Value, payload []byte, sink *int64) error {
	kasCopy := *kasKey
	writer, err := tdf.NewWriter(ctx, tdf.WithDefaultKASForWriter(&kasCopy))
	if err != nil {
		return err
	}

	if _, err := writer.WriteSegment(ctx, 0, payload); err != nil {
		return err
	}

	result, err := writer.Finalize(ctx, tdf.WithAttributeValues(attrs))
	if err != nil {
		return err
	}

	*sink += int64(len(result.Data))
	return nil
}
