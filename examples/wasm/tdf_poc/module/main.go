//go:build wasm

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/opentdf/platform/protocol/go/policy"
	"github.com/opentdf/platform/sdk/experimental/tdf"
)

const (
	statusOK             = 0
	statusError          = -1
	statusNoWriter       = -2
	statusBufferTooSmall = -3

	cryptoModeHost   = int32(0)
	cryptoModeModule = int32(1)
)

var (
	writerMu     sync.Mutex
	activeWriter *tdf.Writer
	lastError    string

	pinnedMu      sync.Mutex
	pinnedBuffers = make(map[int32][]byte)

	lastSegmentData      []byte
	lastSegmentInfo      []byte
	lastFinalizeInfo     []byte
	lastFinalizeManifest []byte

	cryptoMode           = cryptoModeHost
	loggingEnabled int32 = 1
)

func moduleLog(format string, args ...interface{}) {
	if atomic.LoadInt32(&loggingEnabled) != 0 {
		log.Printf(format, args...)
	}
}

func main() {
	moduleLog("[wasm] module loaded; call TDFNewWriter/TDFWriteSegment/TDFFinalize (default crypto mode=%s)", cryptoModeLabel(cryptoMode))
}

func setError(err error) int32 {
	if err == nil {
		lastError = ""
		return statusOK
	}
	lastError = err.Error()
	moduleLog("[wasm] error: %v", err)
	return statusError
}

func clearError() {
	lastError = ""
}

func bytesFrom(ptr, length int32) ([]byte, error) {
	if length == 0 {
		return nil, nil
	}
	if ptr == 0 {
		return nil, errors.New("nil pointer with non-zero length")
	}
	data := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), length)
	out := make([]byte, length)
	copy(out, data)
	return out, nil
}

func writeBytes(ptr, cap int32, data []byte) (int32, error) {
	if len(data) == 0 {
		return 0, nil
	}
	if ptr == 0 {
		return 0, errors.New("nil pointer")
	}
	if int(cap) < len(data) {
		return 0, fmt.Errorf("buffer too small: need %d, have %d", len(data), cap)
	}
	buf := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), cap)
	copy(buf, data)
	return int32(len(data)), nil
}

func copyOut(data []byte, ptr, cap int32) int32 {
	length := len(data)
	if ptr == 0 || cap == 0 {
		return int32(length)
	}
	if int(cap) < length {
		return statusBufferTooSmall
	}
	if cap > 0 {
		buf := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), cap)
		copy(buf, data)
		if length < int(cap) {
			buf[length] = 0
		}
	}
	return int32(length)
}

func pinBuffer(buf []byte) int32 {
	ptr := int32(uintptr(unsafe.Pointer(&buf[0])))
	pinnedMu.Lock()
	pinnedBuffers[ptr] = buf
	pinnedMu.Unlock()
	return ptr
}

func unpinBuffer(ptr int32) {
	pinnedMu.Lock()
	delete(pinnedBuffers, ptr)
	pinnedMu.Unlock()
}

func cryptoModeLabel(mode int32) string {
	switch mode {
	case cryptoModeHost:
		return "host"
	case cryptoModeModule:
		return "module"
	default:
		return fmt.Sprintf("unknown(%d)", mode)
	}
}

//go:wasmexport TDFSetCryptoMode
func TDFSetCryptoMode(mode int32) int32 {
	writerMu.Lock()
	defer writerMu.Unlock()

	switch mode {
	case cryptoModeHost, cryptoModeModule:
		cryptoMode = mode
		clearError()
		moduleLog("[wasm] crypto mode set to %s", cryptoModeLabel(mode))
		return statusOK
	default:
		return setError(fmt.Errorf("invalid crypto mode %d", mode))
	}
}

//go:wasmexport TDFGetCryptoMode
func TDFGetCryptoMode() int32 {
	writerMu.Lock()
	defer writerMu.Unlock()
	return cryptoMode
}

//go:wasmexport TDFSetLogEnabled
func TDFSetLogEnabled(flag int32) int32 {
	if flag != 0 {
		atomic.StoreInt32(&loggingEnabled, 1)
	} else {
		atomic.StoreInt32(&loggingEnabled, 0)
	}
	return statusOK
}

//go:wasmexport TDFGetLogEnabled
func TDFGetLogEnabled() int32 {
	if atomic.LoadInt32(&loggingEnabled) != 0 {
		return 1
	}
	return 0
}

//go:wasmexport TDFAlloc
func TDFAlloc(length int32) int32 {
	if length <= 0 {
		return 0
	}
	buf := make([]byte, length)
	return pinBuffer(buf)
}

//go:wasmexport TDFFree
func TDFFree(ptr int32) int32 {
	if ptr == 0 {
		return statusError
	}
	unpinBuffer(ptr)
	return statusOK
}

//go:wasmexport TDFNewWriter
func TDFNewWriter(defaultKasPtr, defaultKasLen int32) int32 {
	writerMu.Lock()
	defer writerMu.Unlock()

	ctx := context.Background()
	opts := []tdf.Option[*tdf.WriterConfig]{}

	if defaultKasLen > 0 {
		raw, err := bytesFrom(defaultKasPtr, defaultKasLen)
		if err != nil {
			return setError(err)
		}
		var kas policy.SimpleKasKey
		if err := json.Unmarshal(raw, &kas); err != nil {
			return setError(err)
		}
		opts = append(opts, tdf.WithDefaultKASForWriter(&kas))
	}

	mode := cryptoMode
	if mode == cryptoModeModule {
		opts = append(opts, tdf.WithCryptoProvider(newModuleCryptoProvider(nil)))
	}

	moduleLog("[wasm] TDFNewWriter using crypto mode=%s", cryptoModeLabel(mode))

	w, err := tdf.NewWriter(ctx, opts...)
	if err != nil {
		return setError(err)
	}

	activeWriter = w
	lastSegmentData = nil
	lastSegmentInfo = nil
	lastFinalizeInfo = nil
	lastFinalizeManifest = nil
	clearError()
	return statusOK
}

//go:wasmexport TDFWriteSegment
func TDFWriteSegment(index int32, dataPtr, dataLen, outPtr, outCap int32) int32 {
	writerMu.Lock()
	defer writerMu.Unlock()

	if activeWriter == nil {
		lastError = "writer not initialized"
		return statusNoWriter
	}

	data, err := bytesFrom(dataPtr, dataLen)
	if err != nil {
		return setError(err)
	}

	segResult, err := activeWriter.WriteSegment(context.Background(), int(index), data)
	if err != nil {
		return setError(err)
	}

	moduleLog("[wasm] segment %d archive bytes=%d", index, len(segResult.Data))

	lastSegmentData = append(lastSegmentData[:0], segResult.Data...)

	if int(outCap) < len(segResult.Data) {
		lastError = fmt.Sprintf("segment output requires %d bytes", len(segResult.Data))
		return statusBufferTooSmall
	}

	if _, err := writeBytes(outPtr, outCap, segResult.Data); err != nil {
		return setError(err)
	}

	segCopy := segResult
	segCopy.Data = nil
	if infoBytes, err := json.Marshal(segCopy); err == nil {
		lastSegmentInfo = infoBytes
	} else {
		moduleLog("[wasm] failed to marshal segment info: %v", err)
		lastSegmentInfo = nil
	}

	clearError()
	return int32(len(segResult.Data))
}

//go:wasmexport TDFFinalize
func TDFFinalize(attrsPtr, attrsLen, outPtr, outCap int32) int32 {
	writerMu.Lock()
	defer writerMu.Unlock()

	if activeWriter == nil {
		lastError = "writer not initialized"
		return statusNoWriter
	}

	var opts []tdf.Option[*tdf.WriterFinalizeConfig]

	if attrsLen > 0 {
		raw, err := bytesFrom(attrsPtr, attrsLen)
		if err != nil {
			return setError(err)
		}
		var attrs []*policy.Value
		if err := json.Unmarshal(raw, &attrs); err != nil {
			return setError(err)
		}
		opts = append(opts, tdf.WithAttributeValues(attrs))
	}

	result, err := activeWriter.Finalize(context.Background(), opts...)
	if err != nil {
		return setError(err)
	}

	if int(outCap) < len(result.Data) {
		lastError = fmt.Sprintf("finalize output requires %d bytes", len(result.Data))
		return statusBufferTooSmall
	}

	if _, err := writeBytes(outPtr, outCap, result.Data); err != nil {
		return setError(err)
	}

	if manifestBytes, err := json.Marshal(result.Manifest); err == nil {
		lastFinalizeManifest = manifestBytes
	} else {
		moduleLog("[wasm] failed to marshal manifest: %v", err)
		lastFinalizeManifest = nil
	}

	infoCopy := *result
	infoCopy.Data = nil
	infoCopy.Manifest = nil
	if infoBytes, err := json.Marshal(infoCopy); err == nil {
		lastFinalizeInfo = infoBytes
	} else {
		moduleLog("[wasm] failed to marshal finalize info: %v", err)
		lastFinalizeInfo = nil
	}

	lastSegmentData = nil
	activeWriter = nil
	clearError()
	return int32(len(result.Data))
}

//go:wasmexport TDFGetLastSegmentInfo
func TDFGetLastSegmentInfo(ptr, cap int32) int32 {
	writerMu.Lock()
	defer writerMu.Unlock()
	return copyOut(lastSegmentInfo, ptr, cap)
}

//go:wasmexport TDFGetLastSegmentData
func TDFGetLastSegmentData(ptr, cap int32) int32 {
	writerMu.Lock()
	defer writerMu.Unlock()
	return copyOut(lastSegmentData, ptr, cap)
}

//go:wasmexport TDFGetFinalizeInfo
func TDFGetFinalizeInfo(ptr, cap int32) int32 {
	writerMu.Lock()
	defer writerMu.Unlock()
	return copyOut(lastFinalizeInfo, ptr, cap)
}

//go:wasmexport TDFGetFinalizeManifest
func TDFGetFinalizeManifest(ptr, cap int32) int32 {
	writerMu.Lock()
	defer writerMu.Unlock()
	return copyOut(lastFinalizeManifest, ptr, cap)
}

//go:wasmexport TDFGetLastError
func TDFGetLastError(ptr, cap int32) int32 {
	msg := lastError
	if ptr == 0 || cap == 0 {
		return int32(len(msg))
	}
	if int(cap) < len(msg) {
		return statusBufferTooSmall
	}
	buf := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), cap)
	copy(buf, msg)
	if len(msg) < int(cap) {
		buf[len(msg)] = 0
	}
	return int32(len(msg))
}

//go:wasmexport TDFReset
func TDFReset() int32 {
	writerMu.Lock()
	defer writerMu.Unlock()
	activeWriter = nil
	lastSegmentData = nil
	lastSegmentInfo = nil
	lastFinalizeInfo = nil
	lastFinalizeManifest = nil
	clearError()
	return statusOK
}
