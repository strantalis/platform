//go:build wasm

// Experimental: This package is EXPERIMENTAL and may change or be removed at any time

package tdf

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strings"
	"unsafe"
)

const (
	hostStatusOK        = 0
	hostFailureCode     = math.MaxUint32
	hostFailureCode64   = math.MaxUint64
	gcmNonceSize        = 12
	gcmTagOverhead      = 16
	gcmPrefixOverhead   = gcmNonceSize + gcmTagOverhead
	hmacSHA256Size      = 32
	maxWrappedKeySize   = 4096
	maxEphemeralKeySize = 4096
)

type wasmCryptoProvider struct {
	entropyOverride EntropySource
}

func newDefaultCryptoProvider(entropy EntropySource) CryptoProvider {
	return &wasmCryptoProvider{entropyOverride: entropy}
}

func (p *wasmCryptoProvider) RandomBytes(length int) ([]byte, error) {
	if length < 0 {
		return nil, fmt.Errorf("invalid random byte length: %d", length)
	}
	if p.entropyOverride != nil {
		return p.entropyOverride.RandomBytes(length)
	}
	buf := make([]byte, length)
	if length == 0 {
		return buf, nil
	}
	status := hostRandomBytes(wasmPtr(buf), uint32(len(buf)))
	if status != hostStatusOK {
		return nil, fmt.Errorf("hostcrypto.random_bytes failed with status %d", status)
	}
	return buf, nil
}

func (p *wasmCryptoProvider) NewAESGCM(key []byte) (AESGCM, error) {
	if len(key) == 0 {
		return nil, errors.New("aes-gcm requires non-empty key")
	}
	dup := append([]byte(nil), key...)
	return &wasmAESGCM{provider: p, key: dup}, nil
}

func (p *wasmCryptoProvider) Base64Encode(data []byte) []byte {
	out := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(out, data)
	return out
}

func (p *wasmCryptoProvider) Base64Decode(data []byte) ([]byte, error) {
	out := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	n, err := base64.StdEncoding.Decode(out, data)
	if err != nil {
		return nil, err
	}
	return out[:n], nil
}

func (p *wasmCryptoProvider) HMACSHA256(key, data []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("hmac requires non-empty key")
	}
	out := make([]byte, hmacSHA256Size)
	status := hostHmacSha256(wasmPtr(key), uint32(len(key)), wasmPtr(data), uint32(len(data)), wasmPtr(out))
	if status != hostStatusOK {
		return nil, fmt.Errorf("hostcrypto.hmac_sha256 failed with status %d", status)
	}
	return out, nil
}

func (p *wasmCryptoProvider) WrapKey(req KeyWrapRequest) (KeyWrapResult, error) {
	if req.PublicKeyPEM == "" {
		return KeyWrapResult{}, errors.New("public key PEM must not be empty")
	}
	if len(req.PlaintextKey) == 0 {
		return KeyWrapResult{}, errors.New("plaintext key must not be empty")
	}
	salt := req.Salt
	if len(salt) == 0 {
		salt = tdfSalt()
	}
	wrappedBuf := make([]byte, maxWrappedKeySize)
	ephemeralBuf := make([]byte, maxEphemeralKeySize)
	algBytes := []byte(req.Algorithm)
	pubKeyBytes := []byte(req.PublicKeyPEM)
	result := hostWrapKey(
		wasmPtr(algBytes), uint32(len(algBytes)),
		wasmPtr(pubKeyBytes), uint32(len(pubKeyBytes)),
		wasmPtr(req.PlaintextKey), uint32(len(req.PlaintextKey)),
		wasmPtr(salt), uint32(len(salt)),
		wasmPtr(wrappedBuf), uint32(len(wrappedBuf)),
		wasmPtr(ephemeralBuf), uint32(len(ephemeralBuf)),
	)
	if result == hostFailureCode64 {
		return KeyWrapResult{}, errors.New("hostcrypto.wrap_key failed")
	}
	wrappedLen := uint32(result & 0xffffffff)
	ephemeralLen := uint32(result >> 32)
	if wrappedLen > uint32(len(wrappedBuf)) {
		return KeyWrapResult{}, errors.New("host returned invalid wrapped key length")
	}
	if ephemeralLen > uint32(len(ephemeralBuf)) {
		return KeyWrapResult{}, errors.New("host returned invalid ephemeral key length")
	}
	res := KeyWrapResult{
		WrappedKey: wrappedBuf[:wrappedLen],
	}
	if strings.HasPrefix(strings.ToLower(req.Algorithm), "ec:") {
		res.Scheme = KeyWrapSchemeEC
	} else {
		res.Scheme = KeyWrapSchemeRSA
	}
	if ephemeralLen > 0 {
		res.EphemeralPublicKey = string(ephemeralBuf[:ephemeralLen])
	}
	return res, nil
}

type wasmAESGCM struct {
	provider *wasmCryptoProvider
	key      []byte
}

func (a *wasmAESGCM) Encrypt(plaintext []byte) ([]byte, error) {
	out := make([]byte, len(plaintext)+gcmPrefixOverhead)
	length := hostAesGcmEncrypt(
		wasmPtr(a.key), uint32(len(a.key)),
		wasmPtr(plaintext), uint32(len(plaintext)),
		wasmPtr(out), uint32(len(out)),
	)
	if length == hostFailureCode {
		return nil, errors.New("hostcrypto.aes_gcm_encrypt failed")
	}
	if length > uint32(len(out)) {
		return nil, errors.New("host returned invalid encrypt length")
	}
	return out[:length], nil
}

func (a *wasmAESGCM) EncryptWithIV(iv, plaintext []byte) ([]byte, error) {
	if len(iv) != gcmNonceSize {
		return nil, fmt.Errorf("aes-gcm requires %d-byte IV", gcmNonceSize)
	}
	out := make([]byte, len(plaintext)+gcmTagOverhead)
	length := hostAesGcmEncryptWithIV(
		wasmPtr(a.key), uint32(len(a.key)),
		wasmPtr(iv), uint32(len(iv)),
		wasmPtr(plaintext), uint32(len(plaintext)),
		wasmPtr(out), uint32(len(out)),
	)
	if length == hostFailureCode {
		return nil, errors.New("hostcrypto.aes_gcm_encrypt_with_iv failed")
	}
	if length > uint32(len(out)) {
		return nil, errors.New("host returned invalid encrypt length")
	}
	return out[:length], nil
}

func (a *wasmAESGCM) EncryptWithIVAndTagSize(iv, plaintext []byte, tagSize int) ([]byte, error) {
	if len(iv) != gcmNonceSize {
		return nil, fmt.Errorf("aes-gcm requires %d-byte IV", gcmNonceSize)
	}
	if tagSize <= 0 || tagSize > gcmTagOverhead {
		return nil, fmt.Errorf("unsupported tag size %d", tagSize)
	}
	out := make([]byte, len(plaintext)+tagSize)
	length := hostAesGcmEncryptWithIVAndTag(
		wasmPtr(a.key), uint32(len(a.key)),
		wasmPtr(iv), uint32(len(iv)),
		wasmPtr(plaintext), uint32(len(plaintext)),
		wasmPtr(out), uint32(len(out)),
		uint32(tagSize),
	)
	if length == hostFailureCode {
		return nil, errors.New("hostcrypto.aes_gcm_encrypt_with_iv_tag failed")
	}
	if length > uint32(len(out)) {
		return nil, errors.New("host returned invalid encrypt length")
	}
	return out[:length], nil
}

func (a *wasmAESGCM) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < gcmPrefixOverhead {
		return nil, errors.New("ciphertext too short for AES-GCM")
	}
	out := make([]byte, len(ciphertext)-gcmPrefixOverhead)
	length := hostAesGcmDecrypt(
		wasmPtr(a.key), uint32(len(a.key)),
		wasmPtr(ciphertext), uint32(len(ciphertext)),
		wasmPtr(out), uint32(len(out)),
	)
	if length == hostFailureCode {
		return nil, errors.New("hostcrypto.aes_gcm_decrypt failed")
	}
	if length > uint32(len(out)) {
		return nil, errors.New("host returned invalid decrypt length")
	}
	return out[:length], nil
}

func (a *wasmAESGCM) DecryptWithTagSize(ciphertext []byte, tagSize int) ([]byte, error) {
	if len(ciphertext) < gcmNonceSize+tagSize {
		return nil, errors.New("ciphertext too short for AES-GCM with custom tag")
	}
	out := make([]byte, len(ciphertext)-(gcmNonceSize+tagSize))
	length := hostAesGcmDecryptWithTag(
		wasmPtr(a.key), uint32(len(a.key)),
		wasmPtr(ciphertext), uint32(len(ciphertext)),
		wasmPtr(out), uint32(len(out)),
		uint32(tagSize),
	)
	if length == hostFailureCode {
		return nil, errors.New("hostcrypto.aes_gcm_decrypt_with_tag failed")
	}
	if length > uint32(len(out)) {
		return nil, errors.New("host returned invalid decrypt length")
	}
	return out[:length], nil
}

func (a *wasmAESGCM) DecryptWithIVAndTagSize(iv, ciphertext []byte, tagSize int) ([]byte, error) {
	if len(iv) != gcmNonceSize {
		return nil, fmt.Errorf("aes-gcm requires %d-byte IV", gcmNonceSize)
	}
	if len(ciphertext) < tagSize {
		return nil, errors.New("ciphertext too short for AES-GCM with IV")
	}
	out := make([]byte, len(ciphertext)-tagSize)
	length := hostAesGcmDecryptWithIVAndTag(
		wasmPtr(a.key), uint32(len(a.key)),
		wasmPtr(iv), uint32(len(iv)),
		wasmPtr(ciphertext), uint32(len(ciphertext)),
		wasmPtr(out), uint32(len(out)),
		uint32(tagSize),
	)
	if length == hostFailureCode {
		return nil, errors.New("hostcrypto.aes_gcm_decrypt_with_iv_tag failed")
	}
	if length > uint32(len(out)) {
		return nil, errors.New("host returned invalid decrypt length")
	}
	return out[:length], nil
}

// wasmPtr returns the linear memory pointer for the slice start. For empty slices, it returns 0.
func wasmPtr(b []byte) uint32 {
	if len(b) == 0 {
		return 0
	}
	return uint32(uintptr(unsafe.Pointer(&b[0])))
}

// Host crypto imports. The host environment must implement these entry points.

//go:wasmimport hostcrypto random_bytes
func hostRandomBytes(ptr uint32, length uint32) uint32

//go:wasmimport hostcrypto hmac_sha256
func hostHmacSha256(keyPtr, keyLen, dataPtr, dataLen, outPtr uint32) uint32

//go:wasmimport hostcrypto aes_gcm_encrypt
func hostAesGcmEncrypt(keyPtr, keyLen, ptPtr, ptLen, outPtr, outCap uint32) uint32

//go:wasmimport hostcrypto aes_gcm_encrypt_with_iv
func hostAesGcmEncryptWithIV(keyPtr, keyLen, ivPtr, ivLen, ptPtr, ptLen, outPtr, outCap uint32) uint32

//go:wasmimport hostcrypto aes_gcm_encrypt_with_iv_tag
func hostAesGcmEncryptWithIVAndTag(keyPtr, keyLen, ivPtr, ivLen, ptPtr, ptLen, outPtr, outCap, tagSize uint32) uint32

//go:wasmimport hostcrypto aes_gcm_decrypt
func hostAesGcmDecrypt(keyPtr, keyLen, ctPtr, ctLen, outPtr, outCap uint32) uint32

//go:wasmimport hostcrypto aes_gcm_decrypt_with_tag
func hostAesGcmDecryptWithTag(keyPtr, keyLen, ctPtr, ctLen, outPtr, outCap, tagSize uint32) uint32

//go:wasmimport hostcrypto aes_gcm_decrypt_with_iv_tag
func hostAesGcmDecryptWithIVAndTag(keyPtr, keyLen, ivPtr, ivLen, ctPtr, ctLen, outPtr, outCap, tagSize uint32) uint32

//go:wasmimport hostcrypto wrap_key
func hostWrapKey(algPtr, algLen, pubPtr, pubLen, keyPtr, keyLen, saltPtr, saltLen, outPtr, outCap, ephPtr, ephCap uint32) uint64
