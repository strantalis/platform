package bench

/*
#include <dlfcn.h>
#include <stdint.h>
#include <stdlib.h>

#ifndef RTLD_LAZY
#define RTLD_LAZY 1
#endif

typedef int32_t (*tdf_new_writer_fn)(const uint8_t*, size_t);
typedef int32_t (*tdf_write_segment_fn)(int32_t, const uint8_t*, size_t, uint8_t*, size_t);
typedef int32_t (*tdf_finalize_fn)(const uint8_t*, size_t, uint8_t*, size_t);
typedef int32_t (*tdf_get_last_error_fn)(char*, size_t);
typedef int32_t (*tdf_reset_fn)(void);

static inline void* go_dlopen(const char* path) {
    return dlopen(path, RTLD_LAZY);
}

static inline void go_dlclose(void* handle) {
    if (handle != NULL) {
        dlclose(handle);
    }
}

static inline void* go_dlsym(void* handle, const char* name) {
    return dlsym(handle, name);
}

static inline const char* go_dlerror(void) {
    return dlerror();
}

static inline int32_t call_TDFNewWriter(void* fn, const uint8_t* kas, size_t len) {
    return ((tdf_new_writer_fn)fn)(kas, len);
}

static inline int32_t call_TDFWriteSegment(void* fn, int32_t index, const uint8_t* data, size_t dataLen, uint8_t* outBuf, size_t outCap) {
    return ((tdf_write_segment_fn)fn)(index, data, dataLen, outBuf, outCap);
}

static inline int32_t call_TDFFinalize(void* fn, const uint8_t* attrs, size_t attrsLen, uint8_t* outBuf, size_t outCap) {
    return ((tdf_finalize_fn)fn)(attrs, attrsLen, outBuf, outCap);
}

static inline int32_t call_TDFGetLastError(void* fn, char* outBuf, size_t outCap) {
    return ((tdf_get_last_error_fn)fn)(outBuf, outCap);
}

static inline int32_t call_TDFReset(void* fn) {
    return ((tdf_reset_fn)fn)();
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	statusOK             = 0
	statusBufferTooSmall = -3
)

type cSharedLib struct {
	handle       unsafe.Pointer
	newWriter    unsafe.Pointer
	writeSegment unsafe.Pointer
	finalize     unsafe.Pointer
	lastError    unsafe.Pointer
	reset        unsafe.Pointer
}

func (lib *cSharedLib) close() {
	if lib.handle != nil {
		C.go_dlclose(lib.handle)
		lib.handle = nil
	}
}

func (lib *cSharedLib) newWriterCall(kas []byte) int32 {
	var kasPtr *C.uint8_t
	if len(kas) > 0 {
		kasPtr = (*C.uint8_t)(unsafe.Pointer(&kas[0]))
	}
	status := C.call_TDFNewWriter(lib.newWriter, kasPtr, C.size_t(len(kas)))
	return int32(status)
}

func (lib *cSharedLib) writeSegmentCall(index int32, payload []byte, out []byte) int32 {
	var payloadPtr *C.uint8_t
	if len(payload) > 0 {
		payloadPtr = (*C.uint8_t)(unsafe.Pointer(&payload[0]))
	}
	var outPtr *C.uint8_t
	var outCap C.size_t
	if len(out) > 0 {
		outPtr = (*C.uint8_t)(unsafe.Pointer(&out[0]))
		outCap = C.size_t(len(out))
	}
	status := C.call_TDFWriteSegment(lib.writeSegment, C.int32_t(index), payloadPtr, C.size_t(len(payload)), outPtr, outCap)
	return int32(status)
}

func (lib *cSharedLib) finalizeCall(attrs []byte, out []byte) int32 {
	var attrsPtr *C.uint8_t
	if len(attrs) > 0 {
		attrsPtr = (*C.uint8_t)(unsafe.Pointer(&attrs[0]))
	}
	var outPtr *C.uint8_t
	var outCap C.size_t
	if len(out) > 0 {
		outPtr = (*C.uint8_t)(unsafe.Pointer(&out[0]))
		outCap = C.size_t(len(out))
	}
	status := C.call_TDFFinalize(lib.finalize, attrsPtr, C.size_t(len(attrs)), outPtr, outCap)
	return int32(status)
}

func (lib *cSharedLib) lastErrorString() string {
	needed := int(C.call_TDFGetLastError(lib.lastError, nil, 0))
	if needed <= 0 {
		return ""
	}
	buf := make([]byte, needed+1)
	status := int(C.call_TDFGetLastError(lib.lastError, (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(len(buf))))
	if status < 0 {
		return ""
	}
	n := 0
	for n < len(buf) && buf[n] != 0 {
		n++
	}
	return string(buf[:n])
}

func (lib *cSharedLib) resetCall() int32 {
	status := C.call_TDFReset(lib.reset)
	return int32(status)
}

func openSharedLib(path string) (*cSharedLib, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	handle := C.go_dlopen(cPath)
	if handle == nil {
		return nil, fmt.Errorf("dlopen failed: %s", dlLastError())
	}
	load := func(name string) (unsafe.Pointer, error) {
		cName := C.CString(name)
		defer C.free(unsafe.Pointer(cName))
		sym := C.go_dlsym(handle, cName)
		if sym == nil {
			return nil, fmt.Errorf("dlsym(%s): %s", name, dlLastError())
		}
		return sym, nil
	}
	newWriter, err := load("TDFNewWriter")
	if err != nil {
		C.go_dlclose(handle)
		return nil, err
	}
	writeSegment, err := load("TDFWriteSegment")
	if err != nil {
		C.go_dlclose(handle)
		return nil, err
	}
	finalize, err := load("TDFFinalize")
	if err != nil {
		C.go_dlclose(handle)
		return nil, err
	}
	lastError, err := load("TDFGetLastError")
	if err != nil {
		C.go_dlclose(handle)
		return nil, err
	}
	reset, err := load("TDFReset")
	if err != nil {
		C.go_dlclose(handle)
		return nil, err
	}
	return &cSharedLib{handle: handle, newWriter: newWriter, writeSegment: writeSegment, finalize: finalize, lastError: lastError, reset: reset}, nil
}

func dlLastError() string {
	errPtr := C.go_dlerror()
	if errPtr == nil {
		return "unknown"
	}
	return C.GoString(errPtr)
}

func callWithBuffer(lib *cSharedLib, op func([]byte) int32) int {
	buf := make([]byte, 64*1024)
	for {
		status := op(buf)
		if status >= 0 {
			return int(status)
		}
		if status == statusBufferTooSmall {
			if len(buf) == 0 {
				buf = make([]byte, 1024)
			} else {
				buf = make([]byte, len(buf)*2)
			}
			continue
		}
		panic(fmt.Sprintf("operation failed: status=%d err=%s", status, lib.lastErrorString()))
	}
}
