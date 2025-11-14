package bench

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func BenchmarkSharedLibraryFFI(b *testing.B) {
	repo := repoRoot(b)
	kas := mustReadFile(b, filepath.Join(repo, "examples/wasm/tdf_poc/cshared/testdata/kas.json"))
	attrs := mustReadFile(b, filepath.Join(repo, "examples/wasm/tdf_poc/cshared/testdata/attributes.json"))
	payload := mustReadFile(b, filepath.Join(repo, "examples/wasm/tdf_poc/cshared/testdata/payload.txt"))

	libPath := buildSharedLibrary(b, repo)
	lib, err := openSharedLib(libPath)
	if err != nil {
		b.Fatalf("open shared lib: %v", err)
	}
	b.Cleanup(lib.close)

	// Warm-up
	runSharedOnce(b, lib, kas, attrs, payload)

	b.ReportAllocs()
	b.ResetTimer()

	var total int
	for i := 0; i < b.N; i++ {
		total += runSharedOnce(b, lib, kas, attrs, payload)
	}
	benchmarkSink = total
}

var benchmarkSink int

func runSharedOnce(b *testing.B, lib *cSharedLib, kas, attrs, payload []byte) int {
	start := time.Now()
	if status := lib.newWriterCall(kas); status != statusOK {
		b.Fatalf("TDFNewWriter failed: status=%d err=%s", status, lib.lastErrorString())
	}

	segment := callWithBuffer(lib, func(buf []byte) int32 {
		return lib.writeSegmentCall(0, payload, buf)
	})

	finalize := callWithBuffer(lib, func(buf []byte) int32 {
		return lib.finalizeCall(attrs, buf)
	})

	if status := lib.resetCall(); status != statusOK {
		b.Fatalf("TDFReset failed: status=%d err=%s", status, lib.lastErrorString())
	}

	elapsed := time.Since(start)
	b.ReportMetric(float64(elapsed.Nanoseconds()), "run_ns/op")
	return segment + finalize
}

func buildSharedLibrary(b *testing.B, repo string) string {
	tmpDir := b.TempDir()
	libPath := filepath.Join(tmpDir, "libtdf.so")
	cmd := exec.Command("go", "build", "-buildmode=c-shared", "-o", libPath, "./examples/wasm/tdf_poc/cshared")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=1",
		fmt.Sprintf("GOCACHE=%s", filepath.Join(repo, ".gocache")),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		b.Fatalf("build c-shared: %v\n%s", err, out)
	}
	return libPath
}

func mustReadFile(b *testing.B, path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatalf("read %s: %v", path, err)
	}
	return data
}

func repoRoot(b *testing.B) string {
	b.Helper()
	dir, err := os.Getwd()
	if err != nil {
		b.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		next := filepath.Dir(dir)
		if next == dir {
			b.Fatalf("could not locate repo root from %s", dir)
		}
		dir = next
	}
}
