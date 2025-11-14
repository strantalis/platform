package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	wasmtime "github.com/bytecodealliance/wasmtime-go/v38"
	"github.com/opentdf/platform/protocol/go/policy"
)

const (
	statusOK             = int32(0)
	statusError          = int32(-1)
	statusNoWriter       = int32(-2)
	statusBufferTooSmall = int32(-3)

	hostStatusOK      = int32(0)
	hostFailureStatus = int32(-1)

	cryptoModeHost   = int32(0)
	cryptoModeModule = int32(1)
)

type namedFunc struct {
	fn   *wasmtime.Func
	name string
}

type wasmRuntime struct {
	engine              *wasmtime.Engine
	instance            *wasmtime.Instance
	store               *wasmtime.Store
	mem                 *wasmtime.Memory
	alloc               namedFunc
	free                namedFunc
	newWriter           namedFunc
	writeSegment        namedFunc
	finalize            namedFunc
	getLastError        namedFunc
	getSegmentInfo      namedFunc
	getSegmentData      namedFunc
	getFinalizeInfo     namedFunc
	getFinalizeManifest namedFunc
	setCryptoMode       namedFunc
	getCryptoMode       namedFunc
	setLogEnabled       namedFunc
	getLogEnabled       namedFunc
}

type modeRequest struct {
	name  string
	value int32
}

func (nf namedFunc) callI32(store *wasmtime.Store, args ...interface{}) int32 {
	res, err := nf.fn.Call(store, args...)
	if err != nil {
		log.Fatalf("call to %s failed: %v", nf.name, err)
	}
	if res == nil {
		return 0
	}
	val, ok := res.(int32)
	if !ok {
		log.Fatalf("%s returned unexpected type %T", nf.name, res)
	}
	return val
}

func (nf namedFunc) callVoid(store *wasmtime.Store, args ...interface{}) {
	if _, err := nf.fn.Call(store, args...); err != nil {
		log.Fatalf("call to %s failed: %v", nf.name, err)
	}
}

func (rt *wasmRuntime) setMode(mode int32) {
	status := rt.setCryptoMode.callI32(rt.store, mode)
	ensureOK(rt.store, rt.mem, rt.alloc, rt.free, rt.getLastError, status, "TDFSetCryptoMode")
}

func (rt *wasmRuntime) mode() int32 {
	return rt.getCryptoMode.callI32(rt.store)
}

func (rt *wasmRuntime) run(kasJSON, attrsJSON, payload []byte, segmentCap int32, verbose bool) []byte {
	logf := func(format string, args ...interface{}) {
		if verbose {
			log.Printf(format, args...)
		}
	}

	store := rt.store
	mem := rt.mem
	newWriter := rt.newWriter
	writeSegment := rt.writeSegment
	finalize := rt.finalize
	getLastError := rt.getLastError
	getSegmentInfo := rt.getSegmentInfo
	getSegmentData := rt.getSegmentData
	getFinalizeInfo := rt.getFinalizeInfo
	getFinalizeManifest := rt.getFinalizeManifest
	alloc := rt.alloc
	free := rt.free

	var zipBuffer bytes.Buffer

	kasPtr := writeBuffer(store, mem, alloc, kasJSON)
	defer freeBuffer(store, free, kasPtr)

	status := newWriter.callI32(store, kasPtr, int32(len(kasJSON)))
	ensureOK(store, mem, alloc, free, getLastError, status, "TDFNewWriter")

	dataPtr := writeBuffer(store, mem, alloc, payload)
	defer freeBuffer(store, free, dataPtr)

	segmentPtr := alloc.callI32(store, segmentCap)
	if segmentPtr == 0 {
		log.Fatalf("failed to allocate segment buffer")
	}
	defer freeBuffer(store, free, segmentPtr)

	status = writeSegment.callI32(store, int32(0), dataPtr, int32(len(payload)), segmentPtr, segmentCap)
	ensureOK(store, mem, alloc, free, getLastError, status, "TDFWriteSegment")

	if status > 0 {
		chunk := readMemory(store, mem, segmentPtr, status)
		logf("Encrypted segment length: %d bytes", len(chunk))
		zipBuffer.Write(chunk)
	} else {
		chunk := fetchBytes(store, mem, alloc, free, getSegmentData)
		if len(chunk) > 0 {
			logf("Encrypted segment length: %d bytes", len(chunk))
			zipBuffer.Write(chunk)
		} else {
			logf("Encrypted segment length: %d bytes", status)
		}
	}

	if segJSON := fetchBytes(store, mem, alloc, free, getSegmentInfo); len(segJSON) > 0 {
		var info struct {
			Index         int    `json:"index"`
			Hash          string `json:"hash"`
			PlaintextSize int64  `json:"plaintextSize"`
			EncryptedSize int64  `json:"encryptedSize"`
			CRC32         uint32 `json:"crc32"`
		}
		if err := json.Unmarshal(segJSON, &info); err == nil {
			logf("Segment %d summary: plaintext=%d encrypted=%d crc=%08x", info.Index, info.PlaintextSize, info.EncryptedSize, info.CRC32)
		} else {
			logf("failed to parse segment info: %v", err)
		}
	}

	attrsPtr := writeBuffer(store, mem, alloc, attrsJSON)
	defer freeBuffer(store, free, attrsPtr)

	const outCap = int32(128 * 1024)
	outPtr := alloc.callI32(store, outCap)
	if outPtr == 0 {
		log.Fatalf("failed to allocate output buffer")
	}
	defer freeBuffer(store, free, outPtr)

	status = finalize.callI32(store, attrsPtr, int32(len(attrsJSON)), outPtr, outCap)
	ensureOK(store, mem, alloc, free, getLastError, status, "TDFFinalize")

	finalBytes := readMemory(store, mem, outPtr, status)
	logf("Finalize produced %d bytes (TDF zip)", len(finalBytes))
	zipBuffer.Write(finalBytes)

	if infoJSON := fetchBytes(store, mem, alloc, free, getFinalizeInfo); len(infoJSON) > 0 {
		var info struct {
			TotalSegments int   `json:"totalSegments"`
			TotalSize     int64 `json:"totalSize"`
			EncryptedSize int64 `json:"encryptedSize"`
		}
		if err := json.Unmarshal(infoJSON, &info); err == nil {
			logf("Finalize summary: segments=%d plaintextBytes=%d encryptedBytes=%d", info.TotalSegments, info.TotalSize, info.EncryptedSize)
		} else {
			logf("failed to parse finalize info: %v", err)
		}
	}

	if manifestJSON := fetchBytes(store, mem, alloc, free, getFinalizeManifest); len(manifestJSON) > 0 {
		logf("Manifest JSON: %s", string(manifestJSON))
	}

	return zipBuffer.Bytes()
}

func (rt *wasmRuntime) setLogging(enabled bool) {
	if rt.setLogEnabled.fn == nil {
		return
	}
	var flag int32
	if enabled {
		flag = 1
	}
	status := rt.setLogEnabled.callI32(rt.store, flag)
	ensureOK(rt.store, rt.mem, rt.alloc, rt.free, rt.getLastError, status, "TDFSetLogEnabled")
}

func main() {
	wasmPath := flag.String("wasm", "writer.wasm", "path to wasm module")
	outPath := flag.String("out", "", "optional path to write finalized TDF zip")
	modeFlag := flag.String("mode", "host", "crypto mode to exercise: host | module | both")
	iterations := flag.Int("iterations", 1, "number of iterations to run per mode")
	flag.Parse()

	rt, cleanup, err := loadWasmRuntime(*wasmPath)
	if err != nil {
		log.Fatalf("failed to load wasm runtime: %v", err)
	}
	defer cleanup()

	kasKey := demoKasKey()
	attrs := demoAttributes(kasKey)
	payload := []byte("hello wasm tdf")

	kasJSON := mustJSON(kasKey)
	attrsJSON := mustJSON(attrs)

	modeRequests := parseModeRequests(*modeFlag)

	if len(modeRequests) == 0 {
		log.Fatalf("no modes to execute (mode=%q)", *modeFlag)
	}

	for _, req := range modeRequests {
		log.Printf("=== crypto mode: %s ===", req.name)
		rt.setMode(req.value)

		reported := rt.mode()
		log.Printf("Module reports crypto mode: %s", cryptoModeName(reported))

		for i := 0; i < *iterations; i++ {
			start := time.Now()
			output := rt.run(kasJSON, attrsJSON, payload, int32(256*1024), true)
			elapsed := time.Since(start)

			log.Printf("Mode=%s iteration=%d duration=%s payload=%d output=%d", req.name, i+1, elapsed, len(payload), len(output))

			if *outPath != "" {
				path := formatOutputPath(*outPath, req.name, i, len(modeRequests), *iterations)
				if err := os.WriteFile(path, output, 0o600); err != nil {
					log.Fatalf("failed to write TDF zip to %s: %v", path, err)
				}
				log.Printf("Wrote finalized TDF to %s", path)
			}
		}
	}
}

func loadWasmRuntime(wasmPath string) (*wasmRuntime, func(), error) {
	engine := wasmtime.NewEngine()
	module, err := wasmtime.NewModuleFromFile(engine, wasmPath)
	if err != nil {
		return nil, nil, fmt.Errorf("compile module: %w", err)
	}

	store := wasmtime.NewStore(engine)
	wasiConfig := wasmtime.NewWasiConfig()
	wasiConfig.InheritStdout()
	wasiConfig.InheritStderr()
	wasiConfig.InheritStdin()
	store.SetWasi(wasiConfig)

	linker := wasmtime.NewLinker(engine)
	if err := linker.DefineWasi(); err != nil {
		return nil, nil, fmt.Errorf("define wasi: %w", err)
	}

	if err := linker.DefineFunc(store, "hostcrypto", "random_bytes", randomBytes); err != nil {
		return nil, nil, fmt.Errorf("define hostcrypto.random_bytes: %w", err)
	}
	if err := linker.DefineFunc(store, "hostcrypto", "hmac_sha256", hmacSHA256); err != nil {
		return nil, nil, fmt.Errorf("define hostcrypto.hmac_sha256: %w", err)
	}
	if err := linker.DefineFunc(store, "hostcrypto", "aes_gcm_encrypt", aesGCMEncrypt); err != nil {
		return nil, nil, fmt.Errorf("define hostcrypto.aes_gcm_encrypt: %w", err)
	}
	if err := linker.DefineFunc(store, "hostcrypto", "aes_gcm_encrypt_with_iv", aesGCMEncryptWithIV); err != nil {
		return nil, nil, fmt.Errorf("define hostcrypto.aes_gcm_encrypt_with_iv: %w", err)
	}
	if err := linker.DefineFunc(store, "hostcrypto", "aes_gcm_encrypt_with_iv_tag", aesGCMEncryptWithIVAndTag); err != nil {
		return nil, nil, fmt.Errorf("define hostcrypto.aes_gcm_encrypt_with_iv_tag: %w", err)
	}
	if err := linker.DefineFunc(store, "hostcrypto", "aes_gcm_decrypt", aesGCMDecrypt); err != nil {
		return nil, nil, fmt.Errorf("define hostcrypto.aes_gcm_decrypt: %w", err)
	}
	if err := linker.DefineFunc(store, "hostcrypto", "aes_gcm_decrypt_with_tag", aesGCMDecryptWithTag); err != nil {
		return nil, nil, fmt.Errorf("define hostcrypto.aes_gcm_decrypt_with_tag: %w", err)
	}
	if err := linker.DefineFunc(store, "hostcrypto", "aes_gcm_decrypt_with_iv_tag", aesGCMDecryptWithIVAndTag); err != nil {
		return nil, nil, fmt.Errorf("define hostcrypto.aes_gcm_decrypt_with_iv_tag: %w", err)
	}
	if err := linker.DefineFunc(store, "hostcrypto", "wrap_key", wrapKey); err != nil {
		return nil, nil, fmt.Errorf("define hostcrypto.wrap_key: %w", err)
	}

	instance, err := linker.Instantiate(store, module)
	if err != nil {
		return nil, nil, fmt.Errorf("instantiate module: %w", err)
	}

	if start := instance.GetFunc(store, "_start"); start != nil {
		if _, err := start.Call(store); err != nil {
			if trap, ok := err.(*wasmtime.Trap); ok {
				if !strings.Contains(trap.Message(), "exit status 0") {
					return nil, nil, fmt.Errorf("_start trap: %s", trap.Message())
				}
			} else if !strings.Contains(err.Error(), "exit status 0") {
				return nil, nil, fmt.Errorf("_start failed: %w", err)
			}
		}
	}

	memExport := instance.GetExport(store, "memory")
	if memExport == nil || memExport.Memory() == nil {
		return nil, nil, fmt.Errorf("memory export not found")
	}
	mem := memExport.Memory()

	rt := &wasmRuntime{
		engine:              engine,
		instance:            instance,
		store:               store,
		mem:                 mem,
		alloc:               named(instance, store, "TDFAlloc"),
		free:                named(instance, store, "TDFFree"),
		newWriter:           named(instance, store, "TDFNewWriter"),
		writeSegment:        named(instance, store, "TDFWriteSegment"),
		finalize:            named(instance, store, "TDFFinalize"),
		getLastError:        named(instance, store, "TDFGetLastError"),
		getSegmentInfo:      named(instance, store, "TDFGetLastSegmentInfo"),
		getSegmentData:      named(instance, store, "TDFGetLastSegmentData"),
		getFinalizeInfo:     named(instance, store, "TDFGetFinalizeInfo"),
		getFinalizeManifest: named(instance, store, "TDFGetFinalizeManifest"),
		setCryptoMode:       named(instance, store, "TDFSetCryptoMode"),
		getCryptoMode:       named(instance, store, "TDFGetCryptoMode"),
	}

	if fn := instance.GetFunc(store, "TDFSetLogEnabled"); fn != nil {
		rt.setLogEnabled = namedFunc{fn: fn, name: "TDFSetLogEnabled"}
	}
	if fn := instance.GetFunc(store, "TDFGetLogEnabled"); fn != nil {
		rt.getLogEnabled = namedFunc{fn: fn, name: "TDFGetLogEnabled"}
	}

	cleanup := func() {
		runtime.KeepAlive(engine)
		runtime.KeepAlive(module)
		runtime.KeepAlive(instance)
	}

	return rt, cleanup, nil
}

func parseModeRequests(mode string) []modeRequest {
	switch strings.ToLower(mode) {
	case "host":
		return []modeRequest{{name: "host", value: cryptoModeHost}}
	case "module":
		return []modeRequest{{name: "module", value: cryptoModeModule}}
	case "both":
		return []modeRequest{
			{name: "host", value: cryptoModeHost},
			{name: "module", value: cryptoModeModule},
		}
	default:
		return nil
	}
}

func cryptoModeName(mode int32) string {
	switch mode {
	case cryptoModeHost:
		return "host"
	case cryptoModeModule:
		return "module"
	default:
		return fmt.Sprintf("unknown(%d)", mode)
	}
}

func formatOutputPath(basePath, mode string, iteration int, modeCount, iterations int) string {
	if modeCount == 1 && iterations == 1 {
		return basePath
	}
	ext := filepath.Ext(basePath)
	prefix := strings.TrimSuffix(basePath, ext)
	return fmt.Sprintf("%s_%s_%d%s", prefix, mode, iteration+1, ext)
}

func named(instance *wasmtime.Instance, store *wasmtime.Store, name string) namedFunc {
	fn := instance.GetFunc(store, name)
	if fn == nil {
		log.Fatalf("%s export not found", name)
	}
	return namedFunc{fn: fn, name: name}
}

func writeBuffer(store *wasmtime.Store, mem *wasmtime.Memory, alloc namedFunc, data []byte) int32 {
	length := len(data)
	if length == 0 {
		return 0
	}
	ptr := alloc.callI32(store, int32(length))
	if ptr == 0 {
		log.Fatalf("failed to allocate %d bytes", length)
	}
	raw := mem.UnsafeData(store)
	if int(ptr) < 0 || int(ptr)+length > len(raw) {
		log.Fatalf("memory write out of bounds: ptr=%d len=%d", ptr, length)
	}
	copy(raw[int(ptr):int(ptr)+length], data)
	runtime.KeepAlive(mem)
	return ptr
}

func readMemory(store *wasmtime.Store, mem *wasmtime.Memory, ptr, length int32) []byte {
	if length <= 0 {
		return nil
	}
	raw := mem.UnsafeData(store)
	if int(ptr) < 0 || int(ptr)+int(length) > len(raw) {
		log.Fatalf("memory read out of bounds: ptr=%d len=%d", ptr, length)
	}
	out := make([]byte, length)
	copy(out, raw[int(ptr):int(ptr)+int(length)])
	runtime.KeepAlive(mem)
	return out
}

func freeBuffer(store *wasmtime.Store, free namedFunc, ptr int32) {
	if ptr == 0 {
		return
	}
	status := free.callI32(store, ptr)
	if status != statusOK {
		log.Printf("TDFFree(%d) returned status %d", ptr, status)
	}
}

func ensureOK(store *wasmtime.Store, mem *wasmtime.Memory, alloc, free, getLastError namedFunc, status int32, op string) {
	if status >= 0 {
		return
	}
	msg := fetchError(store, mem, alloc, free, getLastError)
	log.Fatalf("%s failed (status %d): %s", op, status, msg)
}

func fetchError(store *wasmtime.Store, mem *wasmtime.Memory, alloc, free, getLastError namedFunc) string {
	length := getLastError.callI32(store, int32(0), int32(0))
	if length <= 0 {
		return "unknown error"
	}
	bufPtr := alloc.callI32(store, length+1)
	if bufPtr == 0 {
		return fmt.Sprintf("unable to allocate error buffer (len %d)", length)
	}
	defer freeBuffer(store, free, bufPtr)

	status := getLastError.callI32(store, bufPtr, length+1)
	if status < 0 {
		return fmt.Sprintf("failed to read error message (status %d)", status)
	}
	data := readMemory(store, mem, bufPtr, status)
	return string(data)
}

func fetchBytes(store *wasmtime.Store, mem *wasmtime.Memory, alloc, free, getter namedFunc) []byte {
	length := getter.callI32(store, int32(0), int32(0))
	if length <= 0 {
		return nil
	}
	bufPtr := alloc.callI32(store, length+1)
	if bufPtr == 0 {
		log.Printf("unable to allocate buffer for metadata (len %d)", length)
		return nil
	}
	defer freeBuffer(store, free, bufPtr)

	written := getter.callI32(store, bufPtr, length+1)
	if written < 0 {
		log.Printf("failed to retrieve metadata: status %d", written)
		return nil
	}
	return readMemory(store, mem, bufPtr, written)
}

func demoKasKey() *policy.SimpleKasKey {
	return &policy.SimpleKasKey{
		KasUri: "tdf://kas/demo",
		PublicKey: &policy.SimpleKasPublicKey{
			Algorithm: policy.Algorithm_ALGORITHM_RSA_2048,
			Kid:       "demo-kid",
			Pem:       demoKasPublicKey,
		},
	}
}

func demoAttributes(kasKey *policy.SimpleKasKey) []*policy.Value {
	return []*policy.Value{
		{
			Fqn: "https://example.com/attr/Default/value/Example",
			Attribute: &policy.Attribute{
				Fqn:  "https://example.com/attr/Default",
				Name: "Default",
				Rule: policy.AttributeRuleTypeEnum_ATTRIBUTE_RULE_TYPE_ENUM_ALL_OF,
				Namespace: &policy.Namespace{
					Fqn:  "https://example.com/attr",
					Id:   "example-ns",
					Name: "example.com",
				},
			},
			Value: "Example",
			Grants: []*policy.KeyAccessServer{
				{
					Uri:     "tdf://kas/demo",
					KasKeys: []*policy.SimpleKasKey{kasKey},
				},
			},
		},
	}
}

func mustJSON(value interface{}) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		log.Fatalf("failed to marshal value: %v", err)
	}
	return data
}

func memory(caller *wasmtime.Caller) *wasmtime.Memory {
	export := caller.GetExport("memory")
	if export == nil {
		panic("memory export not found")
	}
	mem := export.Memory()
	if mem == nil {
		panic("export is not memory")
	}
	return mem
}

func readBytes(caller *wasmtime.Caller, ptr, length int32) ([]byte, error) {
	if ptr < 0 || length < 0 {
		return nil, fmt.Errorf("invalid pointer (%d) or length (%d)", ptr, length)
	}
	if length == 0 {
		return []byte{}, nil
	}
	mem := memory(caller)
	raw := mem.UnsafeData(caller)
	if int(ptr)+int(length) > len(raw) {
		return nil, fmt.Errorf("memory read out of bounds: ptr=%d len=%d", ptr, length)
	}
	buf := make([]byte, length)
	copy(buf, raw[int(ptr):int(ptr)+int(length)])
	runtime.KeepAlive(mem)
	return buf, nil
}

func writeBytes(caller *wasmtime.Caller, ptr, capacity int32, data []byte) error {
	if ptr < 0 || capacity < 0 {
		return fmt.Errorf("invalid pointer (%d) or capacity (%d)", ptr, capacity)
	}
	if len(data) > int(capacity) {
		return fmt.Errorf("buffer too small: need %d, have %d", len(data), capacity)
	}
	if len(data) == 0 {
		return nil
	}
	mem := memory(caller)
	raw := mem.UnsafeData(caller)
	if int(ptr)+len(data) > len(raw) {
		return fmt.Errorf("memory write out of bounds: ptr=%d len=%d", ptr, len(data))
	}
	copy(raw[int(ptr):int(ptr)+len(data)], data)
	runtime.KeepAlive(mem)
	return nil
}

func randomBytes(caller *wasmtime.Caller, ptr, length int32) int32 {
	log.Printf("hostcrypto.random_bytes len=%d", length)
	if length < 0 {
		return hostFailureStatus
	}
	if length == 0 {
		return hostStatusOK
	}
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		log.Printf("random_bytes failed: %v", err)
		return hostFailureStatus
	}
	if err := writeBytes(caller, ptr, length, buf); err != nil {
		log.Printf("random_bytes write failed: %v", err)
		return hostFailureStatus
	}
	return hostStatusOK
}

func hmacSHA256(caller *wasmtime.Caller, keyPtr, keyLen, dataPtr, dataLen, outPtr int32) int32 {
	log.Printf("hostcrypto.hmac_sha256 keyLen=%d dataLen=%d", keyLen, dataLen)
	key, err := readBytes(caller, keyPtr, keyLen)
	if err != nil {
		return hostFailureStatus
	}
	data, err := readBytes(caller, dataPtr, dataLen)
	if err != nil {
		return hostFailureStatus
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	sum := mac.Sum(nil)
	if err := writeBytes(caller, outPtr, 32, sum); err != nil {
		return hostFailureStatus
	}
	return hostStatusOK
}

func aesGCMEncrypt(caller *wasmtime.Caller, keyPtr, keyLen, ptPtr, ptLen, outPtr, outCap int32) int32 {
	log.Printf("hostcrypto.aes_gcm_encrypt ptLen=%d", ptLen)
	key, err := readBytes(caller, keyPtr, keyLen)
	if err != nil {
		return hostFailureStatus
	}
	plaintext, err := readBytes(caller, ptPtr, ptLen)
	if err != nil {
		return hostFailureStatus
	}
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		log.Printf("aes_gcm_encrypt nonce error: %v", err)
		return hostFailureStatus
	}
	cipherText, err := encryptAESGCM(key, nonce, plaintext, 16)
	if err != nil {
		return hostFailureStatus
	}
	out := make([]byte, len(nonce)+len(cipherText))
	copy(out, nonce)
	copy(out[len(nonce):], cipherText)
	if err := writeBytes(caller, outPtr, outCap, out); err != nil {
		return hostFailureStatus
	}
	return int32(len(out))
}

func aesGCMEncryptWithIV(caller *wasmtime.Caller, keyPtr, keyLen, ivPtr, ivLen, ptPtr, ptLen, outPtr, outCap int32) int32 {
	log.Printf("hostcrypto.aes_gcm_encrypt_with_iv ptLen=%d", ptLen)
	key, err := readBytes(caller, keyPtr, keyLen)
	if err != nil {
		return hostFailureStatus
	}
	iv, err := readBytes(caller, ivPtr, ivLen)
	if err != nil || len(iv) != 12 {
		return hostFailureStatus
	}
	plaintext, err := readBytes(caller, ptPtr, ptLen)
	if err != nil {
		return hostFailureStatus
	}
	cipherText, err := encryptAESGCM(key, iv, plaintext, 16)
	if err != nil {
		return hostFailureStatus
	}
	if err := writeBytes(caller, outPtr, outCap, cipherText); err != nil {
		return hostFailureStatus
	}
	return int32(len(cipherText))
}

func aesGCMEncryptWithIVAndTag(caller *wasmtime.Caller, keyPtr, keyLen, ivPtr, ivLen, ptPtr, ptLen, outPtr, outCap, tagSize int32) int32 {
	log.Printf("hostcrypto.aes_gcm_encrypt_with_iv_tag ptLen=%d tagSize=%d", ptLen, tagSize)
	if tagSize <= 0 || tagSize > 16 {
		return hostFailureStatus
	}
	key, err := readBytes(caller, keyPtr, keyLen)
	if err != nil {
		return hostFailureStatus
	}
	iv, err := readBytes(caller, ivPtr, ivLen)
	if err != nil || len(iv) != 12 {
		return hostFailureStatus
	}
	plaintext, err := readBytes(caller, ptPtr, ptLen)
	if err != nil {
		return hostFailureStatus
	}
	cipherText, err := encryptAESGCM(key, iv, plaintext, int(tagSize))
	if err != nil {
		return hostFailureStatus
	}
	if err := writeBytes(caller, outPtr, outCap, cipherText); err != nil {
		return hostFailureStatus
	}
	return int32(len(cipherText))
}

func aesGCMDecrypt(caller *wasmtime.Caller, keyPtr, keyLen, ctPtr, ctLen, outPtr, outCap int32) int32 {
	log.Printf("hostcrypto.aes_gcm_decrypt ctLen=%d", ctLen)
	key, err := readBytes(caller, keyPtr, keyLen)
	if err != nil {
		return hostFailureStatus
	}
	ciphertext, err := readBytes(caller, ctPtr, ctLen)
	if err != nil || len(ciphertext) < 12+16 {
		return hostFailureStatus
	}
	nonce := ciphertext[:12]
	body := ciphertext[12:]
	plain, err := decryptAESGCM(key, nonce, body, 16)
	if err != nil {
		return hostFailureStatus
	}
	if err := writeBytes(caller, outPtr, outCap, plain); err != nil {
		return hostFailureStatus
	}
	return int32(len(plain))
}

func aesGCMDecryptWithTag(caller *wasmtime.Caller, keyPtr, keyLen, ctPtr, ctLen, outPtr, outCap, tagSize int32) int32 {
	log.Printf("hostcrypto.aes_gcm_decrypt_with_tag ctLen=%d tagSize=%d", ctLen, tagSize)
	if tagSize <= 0 || tagSize > 16 {
		return hostFailureStatus
	}
	key, err := readBytes(caller, keyPtr, keyLen)
	if err != nil {
		return hostFailureStatus
	}
	ciphertext, err := readBytes(caller, ctPtr, ctLen)
	if err != nil {
		return hostFailureStatus
	}
	nonce := make([]byte, 12)
	plain, err := decryptAESGCM(key, nonce, ciphertext, int(tagSize))
	if err != nil {
		return hostFailureStatus
	}
	if err := writeBytes(caller, outPtr, outCap, plain); err != nil {
		return hostFailureStatus
	}
	return int32(len(plain))
}

func aesGCMDecryptWithIVAndTag(caller *wasmtime.Caller, keyPtr, keyLen, ivPtr, ivLen, ctPtr, ctLen, outPtr, outCap, tagSize int32) int32 {
	log.Printf("hostcrypto.aes_gcm_decrypt_with_iv_tag ctLen=%d tagSize=%d", ctLen, tagSize)
	if tagSize <= 0 || tagSize > 16 {
		return hostFailureStatus
	}
	key, err := readBytes(caller, keyPtr, keyLen)
	if err != nil {
		return hostFailureStatus
	}
	iv, err := readBytes(caller, ivPtr, ivLen)
	if err != nil || len(iv) != 12 {
		return hostFailureStatus
	}
	ciphertext, err := readBytes(caller, ctPtr, ctLen)
	if err != nil {
		return hostFailureStatus
	}
	plain, err := decryptAESGCM(key, iv, ciphertext, int(tagSize))
	if err != nil {
		return hostFailureStatus
	}
	if err := writeBytes(caller, outPtr, outCap, plain); err != nil {
		return hostFailureStatus
	}
	return int32(len(plain))
}

func encryptAESGCM(key, nonce, plaintext []byte, tagSize int) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	var gcm cipher.AEAD
	if tagSize == 16 {
		gcm, err = cipher.NewGCM(block)
	} else {
		gcm, err = cipher.NewGCMWithTagSize(block, tagSize)
	}
	if err != nil {
		return nil, err
	}
	return gcm.Seal(nil, nonce, plaintext, nil), nil
}

func decryptAESGCM(key, nonce, ciphertext []byte, tagSize int) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCMWithTagSize(block, tagSize)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func wrapKey(caller *wasmtime.Caller, algPtr, algLen, pubPtr, pubLen, keyPtr, keyLen, saltPtr, saltLen, outPtr, outCap, ephPtr, ephCap int32) int64 {
	algBytes, err := readBytes(caller, algPtr, algLen)
	if err != nil {
		return failure64()
	}
	alg := string(algBytes)
	log.Printf("hostcrypto.wrap_key algorithm=%s keyLen=%d", alg, keyLen)

	pubBytes, err := readBytes(caller, pubPtr, pubLen)
	if err != nil {
		return failure64()
	}
	keyBytes, err := readBytes(caller, keyPtr, keyLen)
	if err != nil {
		return failure64()
	}

	pemBlock, _ := pem.Decode(pubBytes)
	if pemBlock == nil {
		return failure64()
	}
	pub, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return failure64()
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		log.Printf("wrap key unsupported algorithm: %s", alg)
		return failure64()
	}
	cipherText, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, rsaPub, keyBytes, nil)
	if err != nil {
		return failure64()
	}
	if err := writeBytes(caller, outPtr, outCap, cipherText); err != nil {
		return failure64()
	}
	return encodeWrapResult(len(cipherText), 0)
}

func encodeWrapResult(wrappedLen, ephemeralLen int) int64 {
	return int64(uint64(uint32(ephemeralLen))<<32 | uint64(uint32(wrappedLen)))
}

func failure64() int64 {
	return encodeWrapResult(int(hostFailureStatus), int(hostFailureStatus))
}

const demoKasPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtQ2ZuyT/p32SFmWTj+wQ
huQwR4IJSzlJ7CqZ4fOXw90rA2joK27dIGiHrtkQHGhS4SK1mvkYyJaREoppMFRc
AyZWCgixbSdwYJS/KN0hjLIdhtkdBlZDaZN2ayTf2sZjWzOLL2cYzzVsAy9tGL8a
bMqf91DEHv+l58fPxmbJ/i6YFFQoOEsyWnPhXdiExe6poQDCHJFYYOp6iu5kOPWr
jKFj9eGXuFR/CJQ/uxTSM+8/7Ejmi8Oa52TQAUhMPH0U1CRFm/NuiFoFissa0jJC
J3k6syxvf45mPrbtlhcELskXrquDtJOpIMQmEwfuV4j8iLNwVlsR2tAbClJi6UOy
SQIDAQAB
-----END PUBLIC KEY-----`
