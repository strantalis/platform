package main

import (
	"crypto/sha256"
	"fmt"

	"github.com/opentdf/platform/lib/ocrypto"
	"github.com/opentdf/platform/sdk/experimental/tdf"
)

// moduleCryptoProvider mirrors the default Go crypto provider but is embedded in the WASM module
// so we can benchmark in-module cryptography without host calls.
type moduleCryptoProvider struct {
	entropyOverride tdf.EntropySource
}

func newModuleCryptoProvider(entropy tdf.EntropySource) tdf.CryptoProvider {
	return &moduleCryptoProvider{entropyOverride: entropy}
}

func (p *moduleCryptoProvider) RandomBytes(length int) ([]byte, error) {
	if p.entropyOverride != nil {
		return p.entropyOverride.RandomBytes(length)
	}
	return ocrypto.RandomBytes(length)
}

func (p *moduleCryptoProvider) NewAESGCM(key []byte) (tdf.AESGCM, error) {
	block, err := ocrypto.NewAESGcm(key)
	if err != nil {
		return nil, err
	}
	return &moduleAESGCM{inner: block}, nil
}

func (p *moduleCryptoProvider) Base64Encode(data []byte) []byte {
	return ocrypto.Base64Encode(data)
}

func (p *moduleCryptoProvider) Base64Decode(data []byte) ([]byte, error) {
	return ocrypto.Base64Decode(data)
}

func (p *moduleCryptoProvider) HMACSHA256(key, data []byte) ([]byte, error) {
	return ocrypto.CalculateSHA256Hmac(key, data), nil
}

func (p *moduleCryptoProvider) WrapKey(req tdf.KeyWrapRequest) (tdf.KeyWrapResult, error) {
	keyType := ocrypto.KeyType(req.Algorithm)

	if ocrypto.IsECKeyType(keyType) {
		return p.wrapKeyWithEC(keyType, req)
	}

	if ocrypto.IsRSAKeyType(keyType) {
		return p.wrapKeyWithRSA(req)
	}

	return tdf.KeyWrapResult{}, fmt.Errorf("unsupported key algorithm: %s", req.Algorithm)
}

func (p *moduleCryptoProvider) wrapKeyWithEC(keyType ocrypto.KeyType, req tdf.KeyWrapRequest) (tdf.KeyWrapResult, error) {
	mode, err := ocrypto.ECKeyTypeToMode(keyType)
	if err != nil {
		return tdf.KeyWrapResult{}, fmt.Errorf("failed to derive EC mode: %w", err)
	}

	ecKeyPair, err := ocrypto.NewECKeyPair(mode)
	if err != nil {
		return tdf.KeyWrapResult{}, fmt.Errorf("failed to create EC key pair: %w", err)
	}

	ephemeralPubKey, err := ecKeyPair.PublicKeyInPemFormat()
	if err != nil {
		return tdf.KeyWrapResult{}, fmt.Errorf("failed to encode ephemeral public key: %w", err)
	}

	ephemeralPrivKey, err := ecKeyPair.PrivateKeyInPemFormat()
	if err != nil {
		return tdf.KeyWrapResult{}, fmt.Errorf("failed to encode ephemeral private key: %w", err)
	}

	ecdhKey, err := ocrypto.ComputeECDHKey([]byte(ephemeralPrivKey), []byte(req.PublicKeyPEM))
	if err != nil {
		return tdf.KeyWrapResult{}, fmt.Errorf("failed to compute ECDH key: %w", err)
	}

	salt := req.Salt
	if len(salt) == 0 {
		salt = moduleTDFSalt()
	}

	wrapKey, err := ocrypto.CalculateHKDF(salt, ecdhKey)
	if err != nil {
		return tdf.KeyWrapResult{}, fmt.Errorf("failed to derive wrap key: %w", err)
	}

	if len(wrapKey) > len(req.PlaintextKey) {
		wrapKey = wrapKey[:len(req.PlaintextKey)]
	} else if len(wrapKey) < len(req.PlaintextKey) {
		return tdf.KeyWrapResult{}, fmt.Errorf("wrap key too short: got %d, expected at least %d", len(wrapKey), len(req.PlaintextKey))
	}

	wrapped := make([]byte, len(req.PlaintextKey))
	for i := range req.PlaintextKey {
		wrapped[i] = req.PlaintextKey[i] ^ wrapKey[i]
	}

	return tdf.KeyWrapResult{
		WrappedKey:         wrapped,
		Scheme:             tdf.KeyWrapSchemeEC,
		EphemeralPublicKey: ephemeralPubKey,
	}, nil
}

func (p *moduleCryptoProvider) wrapKeyWithRSA(req tdf.KeyWrapRequest) (tdf.KeyWrapResult, error) {
	encryptor, err := ocrypto.FromPublicPEM(req.PublicKeyPEM)
	if err != nil {
		return tdf.KeyWrapResult{}, fmt.Errorf("failed to create RSA encryptor: %w", err)
	}

	wrapped, err := encryptor.Encrypt(req.PlaintextKey)
	if err != nil {
		return tdf.KeyWrapResult{}, fmt.Errorf("failed to RSA encrypt key: %w", err)
	}

	return tdf.KeyWrapResult{
		WrappedKey: wrapped,
		Scheme:     tdf.KeyWrapSchemeRSA,
	}, nil
}

type moduleAESGCM struct {
	inner ocrypto.AesGcm
}

func (a *moduleAESGCM) Encrypt(plaintext []byte) ([]byte, error) {
	return a.inner.Encrypt(plaintext)
}

func (a *moduleAESGCM) EncryptWithIV(iv, plaintext []byte) ([]byte, error) {
	return a.inner.EncryptWithIV(iv, plaintext)
}

func (a *moduleAESGCM) EncryptWithIVAndTagSize(iv, plaintext []byte, tagSize int) ([]byte, error) {
	return a.inner.EncryptWithIVAndTagSize(iv, plaintext, tagSize)
}

func (a *moduleAESGCM) Decrypt(ciphertext []byte) ([]byte, error) {
	return a.inner.Decrypt(ciphertext)
}

func (a *moduleAESGCM) DecryptWithTagSize(ciphertext []byte, tagSize int) ([]byte, error) {
	return a.inner.DecryptWithTagSize(ciphertext, tagSize)
}

func (a *moduleAESGCM) DecryptWithIVAndTagSize(iv, ciphertext []byte, tagSize int) ([]byte, error) {
	return a.inner.DecryptWithIVAndTagSize(iv, ciphertext, tagSize)
}

var moduleSaltBytes = func() []byte {
	sum := sha256.Sum256([]byte("TDF"))
	out := make([]byte, len(sum))
	copy(out, sum[:])
	return out
}()

func moduleTDFSalt() []byte {
	return append([]byte(nil), moduleSaltBytes...)
}
