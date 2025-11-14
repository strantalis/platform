package main

/*
#cgo CFLAGS: -std=c11
#include <stdint.h>
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"unsafe"

	"github.com/opentdf/platform/protocol/go/policy"
	"github.com/opentdf/platform/sdk/experimental/tdf"
)

const (
	statusOK             = int32(0)
	statusError          = int32(-1)
	statusNoWriter       = int32(-2)
	statusBufferTooSmall = int32(-3)
)

var (
	writerMu             sync.Mutex
	activeWriter         *tdf.Writer
	lastSegmentData      []byte
	lastSegmentInfo      []byte
	lastFinalizeInfo     []byte
	lastFinalizeManifest []byte
	lastError            string
)

func main() {}

//export TDFNewWriter
func TDFNewWriter(kasJSON *C.uint8_t, kasLen C.size_t) C.int32_t {
	writerMu.Lock()
	defer writerMu.Unlock()

	opts := []tdf.Option[*tdf.WriterConfig]{}

	if kasLen > 0 {
		raw, err := bytesFrom(kasJSON, kasLen)
		if err != nil {
			return recordError(err)
		}
		kas := new(policy.SimpleKasKey)
		if err := json.Unmarshal(raw, kas); err != nil {
			return recordError(fmt.Errorf("decode kas json: %w", err))
		}
		opts = append(opts, tdf.WithDefaultKASForWriter(kas))
	}

	writer, err := tdf.NewWriter(context.Background(), opts...)
	if err != nil {
		return recordError(fmt.Errorf("new writer: %w", err))
	}

	activeWriter = writer
	lastSegmentData = nil
	lastSegmentInfo = nil
	lastFinalizeInfo = nil
	lastFinalizeManifest = nil
	clearError()
	return C.int32_t(statusOK)
}

//export TDFWriteSegment
func TDFWriteSegment(index C.int32_t, dataPtr *C.uint8_t, dataLen C.size_t, outPtr *C.uint8_t, outCap C.size_t) C.int32_t {
	writerMu.Lock()
	defer writerMu.Unlock()

	if activeWriter == nil {
		lastError = "writer not initialized"
		return C.int32_t(statusNoWriter)
	}

	payload, err := bytesFrom(dataPtr, dataLen)
	if err != nil {
		return recordError(err)
	}

	segResult, err := activeWriter.WriteSegment(context.Background(), int(index), payload)
	if err != nil {
		return recordError(fmt.Errorf("write segment: %w", err))
	}

	if segResult != nil {
		lastSegmentData = append(lastSegmentData[:0], segResult.Data...)
		segCopy := *segResult
		segCopy.Data = nil
		if infoBytes, err := json.Marshal(segCopy); err == nil {
			lastSegmentInfo = infoBytes
		} else {
			lastSegmentInfo = nil
		}
	} else {
		lastSegmentData = nil
		lastSegmentInfo = nil
	}

	if segResult == nil {
		clearError()
		return C.int32_t(0)
	}

	status := writeOutput(outPtr, outCap, segResult.Data)
	if status != statusOK {
		return C.int32_t(status)
	}

	clearError()
	return C.int32_t(len(segResult.Data))
}

//export TDFFinalize
func TDFFinalize(attrsPtr *C.uint8_t, attrsLen C.size_t, outPtr *C.uint8_t, outCap C.size_t) C.int32_t {
	writerMu.Lock()
	defer writerMu.Unlock()

	if activeWriter == nil {
		lastError = "writer not initialized"
		return C.int32_t(statusNoWriter)
	}

	var finalizeOpts []tdf.Option[*tdf.WriterFinalizeConfig]
	if attrsLen > 0 {
		raw, err := bytesFrom(attrsPtr, attrsLen)
		if err != nil {
			return recordError(err)
		}
		var attrs []*policy.Value
		if err := json.Unmarshal(raw, &attrs); err != nil {
			return recordError(fmt.Errorf("decode attributes json: %w", err))
		}
		finalizeOpts = append(finalizeOpts, tdf.WithAttributeValues(attrs))
	}

	result, err := activeWriter.Finalize(context.Background(), finalizeOpts...)
	if err != nil {
		return recordError(fmt.Errorf("finalize: %w", err))
	}

	if result != nil {
		if manifestBytes, err := json.Marshal(result.Manifest); err == nil {
			lastFinalizeManifest = manifestBytes
		} else {
			lastFinalizeManifest = nil
		}

		infoCopy := *result
		infoCopy.Data = nil
		infoCopy.Manifest = nil
		if infoBytes, err := json.Marshal(infoCopy); err == nil {
			lastFinalizeInfo = infoBytes
		} else {
			lastFinalizeInfo = nil
		}
	} else {
		lastFinalizeInfo = nil
		lastFinalizeManifest = nil
	}

	status := writeOutput(outPtr, outCap, result.Data)
	if status != statusOK {
		return C.int32_t(status)
	}

	activeWriter = nil
	lastSegmentData = nil
	clearError()
	return C.int32_t(len(result.Data))
}

//export TDFGetLastSegmentInfo
func TDFGetLastSegmentInfo(outPtr *C.uint8_t, outCap C.size_t) C.int32_t {
	writerMu.Lock()
	defer writerMu.Unlock()
	return copyOut(lastSegmentInfo, outPtr, outCap)
}

//export TDFGetLastSegmentData
func TDFGetLastSegmentData(outPtr *C.uint8_t, outCap C.size_t) C.int32_t {
	writerMu.Lock()
	defer writerMu.Unlock()
	return copyOut(lastSegmentData, outPtr, outCap)
}

//export TDFGetFinalizeInfo
func TDFGetFinalizeInfo(outPtr *C.uint8_t, outCap C.size_t) C.int32_t {
	writerMu.Lock()
	defer writerMu.Unlock()
	return copyOut(lastFinalizeInfo, outPtr, outCap)
}

//export TDFGetFinalizeManifest
func TDFGetFinalizeManifest(outPtr *C.uint8_t, outCap C.size_t) C.int32_t {
	writerMu.Lock()
	defer writerMu.Unlock()
	return copyOut(lastFinalizeManifest, outPtr, outCap)
}

//export TDFGetLastError
func TDFGetLastError(outPtr *C.char, outCap C.size_t) C.int32_t {
	writerMu.Lock()
	defer writerMu.Unlock()

	if len(lastError) == 0 {
		if outPtr != nil && outCap > 0 {
			buf := unsafe.Slice((*byte)(unsafe.Pointer(outPtr)), int(outCap))
			buf[0] = 0
		}
		return C.int32_t(0)
	}

	needed := len(lastError)
	if outPtr == nil || outCap == 0 {
		return C.int32_t(needed)
	}

	capInt, err := sizeToInt(outCap)
	if err != nil {
		return recordError(err)
	}
	if capInt <= needed {
		return C.int32_t(statusBufferTooSmall)
	}

	buf := unsafe.Slice((*byte)(unsafe.Pointer(outPtr)), capInt)
	copy(buf, lastError)
	buf[needed] = 0
	return C.int32_t(needed)
}

//export TDFReset
func TDFReset() C.int32_t {
	writerMu.Lock()
	defer writerMu.Unlock()
	activeWriter = nil
	lastSegmentData = nil
	lastSegmentInfo = nil
	lastFinalizeInfo = nil
	lastFinalizeManifest = nil
	clearError()
	return C.int32_t(statusOK)
}

func bytesFrom(ptr *C.uint8_t, length C.size_t) ([]byte, error) {
	if length == 0 {
		return nil, nil
	}
	if ptr == nil {
		return nil, errors.New("nil pointer")
	}
	n, err := sizeToInt(length)
	if err != nil {
		return nil, err
	}
	return C.GoBytes(unsafe.Pointer(ptr), C.int(n)), nil
}

func copyOut(data []byte, outPtr *C.uint8_t, outCap C.size_t) C.int32_t {
	length := len(data)
	if outPtr == nil || outCap == 0 {
		return C.int32_t(length)
	}
	capInt, err := sizeToInt(outCap)
	if err != nil {
		return recordError(err)
	}
	if capInt < length {
		return C.int32_t(statusBufferTooSmall)
	}
	if length > 0 {
		buf := unsafe.Slice((*byte)(unsafe.Pointer(outPtr)), capInt)
		copy(buf, data)
		if length < capInt {
			buf[length] = 0
		}
	}
	return C.int32_t(length)
}

func writeOutput(outPtr *C.uint8_t, outCap C.size_t, data []byte) int32 {
	if len(data) == 0 {
		return statusOK
	}
	if outPtr == nil || outCap == 0 {
		lastError = fmt.Sprintf("output requires %d bytes", len(data))
		return statusBufferTooSmall
	}
	capInt, err := sizeToInt(outCap)
	if err != nil {
		lastError = err.Error()
		return statusError
	}
	if capInt < len(data) {
		lastError = fmt.Sprintf("output requires %d bytes", len(data))
		return statusBufferTooSmall
	}
	buf := unsafe.Slice((*byte)(unsafe.Pointer(outPtr)), capInt)
	copy(buf, data)
	return statusOK
}

func sizeToInt(value C.size_t) (int, error) {
	if value > C.size_t(math.MaxInt32) {
		return 0, fmt.Errorf("length %d exceeds supported size", value)
	}
	return int(value), nil
}

func recordError(err error) C.int32_t {
	if err == nil {
		lastError = ""
		return C.int32_t(statusOK)
	}
	lastError = err.Error()
	return C.int32_t(statusError)
}

func clearError() {
	lastError = ""
}
