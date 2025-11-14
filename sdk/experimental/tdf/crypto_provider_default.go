//go:build !wasm

// Experimental: This package is EXPERIMENTAL and may change or be removed at any time

package tdf

import (
	"fmt"

	"github.com/opentdf/platform/lib/ocrypto"
)

// defaultCryptoProvider implements CryptoProvider using the existing Go crypto
// helpers in lib/ocrypto.
type defaultCryptoProvider struct {
	entropyOverride EntropySource
}

func newDefaultCryptoProvider(entropy EntropySource) CryptoProvider {
	return &defaultCryptoProvider{entropyOverride: entropy}
}

func (p *defaultCryptoProvider) RandomBytes(length int) ([]byte, error) {
	if p.entropyOverride != nil {
		return p.entropyOverride.RandomBytes(length)
	}
	return ocrypto.RandomBytes(length)
}

func (p *defaultCryptoProvider) NewAESGCM(key []byte) (AESGCM, error) {
	block, err := ocrypto.NewAESGcm(key)
	if err != nil {
		return nil, err
	}
	return &defaultAESGCM{inner: block}, nil
}

func (p *defaultCryptoProvider) Base64Encode(data []byte) []byte {
	return ocrypto.Base64Encode(data)
}

func (p *defaultCryptoProvider) Base64Decode(data []byte) ([]byte, error) {
	return ocrypto.Base64Decode(data)
}

func (p *defaultCryptoProvider) HMACSHA256(key, data []byte) ([]byte, error) {
	return ocrypto.CalculateSHA256Hmac(key, data), nil
}

func (p *defaultCryptoProvider) WrapKey(req KeyWrapRequest) (KeyWrapResult, error) {
	keyType := ocrypto.KeyType(req.Algorithm)

	if ocrypto.IsECKeyType(keyType) {
		return p.wrapKeyWithEC(keyType, req)
	}

	if ocrypto.IsRSAKeyType(keyType) {
		return p.wrapKeyWithRSA(req)
	}

	return KeyWrapResult{}, fmt.Errorf("unsupported key algorithm: %s", req.Algorithm)
}

type defaultAESGCM struct {
	inner ocrypto.AesGcm
}

func (a *defaultAESGCM) Encrypt(plaintext []byte) ([]byte, error) {
	return a.inner.Encrypt(plaintext)
}

func (a *defaultAESGCM) EncryptWithIV(iv, plaintext []byte) ([]byte, error) {
	return a.inner.EncryptWithIV(iv, plaintext)
}

func (a *defaultAESGCM) EncryptWithIVAndTagSize(iv, plaintext []byte, tagSize int) ([]byte, error) {
	return a.inner.EncryptWithIVAndTagSize(iv, plaintext, tagSize)
}

func (a *defaultAESGCM) Decrypt(ciphertext []byte) ([]byte, error) {
	return a.inner.Decrypt(ciphertext)
}

func (a *defaultAESGCM) DecryptWithTagSize(ciphertext []byte, tagSize int) ([]byte, error) {
	return a.inner.DecryptWithTagSize(ciphertext, tagSize)
}

func (a *defaultAESGCM) DecryptWithIVAndTagSize(iv, ciphertext []byte, tagSize int) ([]byte, error) {
	return a.inner.DecryptWithIVAndTagSize(iv, ciphertext, tagSize)
}

func (p *defaultCryptoProvider) wrapKeyWithEC(keyType ocrypto.KeyType, req KeyWrapRequest) (KeyWrapResult, error) {
	mode, err := ocrypto.ECKeyTypeToMode(keyType)
	if err != nil {
		return KeyWrapResult{}, fmt.Errorf("failed to derive EC mode: %w", err)
	}

	ecKeyPair, err := ocrypto.NewECKeyPair(mode)
	if err != nil {
		return KeyWrapResult{}, fmt.Errorf("failed to create EC key pair: %w", err)
	}

	ephemeralPubKey, err := ecKeyPair.PublicKeyInPemFormat()
	if err != nil {
		return KeyWrapResult{}, fmt.Errorf("failed to encode ephemeral public key: %w", err)
	}

	ephemeralPrivKey, err := ecKeyPair.PrivateKeyInPemFormat()
	if err != nil {
		return KeyWrapResult{}, fmt.Errorf("failed to encode ephemeral private key: %w", err)
	}

	ecdhKey, err := ocrypto.ComputeECDHKey([]byte(ephemeralPrivKey), []byte(req.PublicKeyPEM))
	if err != nil {
		return KeyWrapResult{}, fmt.Errorf("failed to compute ECDH key: %w", err)
	}

	salt := req.Salt
	if len(salt) == 0 {
		salt = tdfSalt()
	}

	wrapKey, err := ocrypto.CalculateHKDF(salt, ecdhKey)
	if err != nil {
		return KeyWrapResult{}, fmt.Errorf("failed to derive wrap key: %w", err)
	}

	if len(wrapKey) > len(req.PlaintextKey) {
		wrapKey = wrapKey[:len(req.PlaintextKey)]
	} else if len(wrapKey) < len(req.PlaintextKey) {
		return KeyWrapResult{}, fmt.Errorf("wrap key too short: got %d, expected at least %d", len(wrapKey), len(req.PlaintextKey))
	}

	wrapped := make([]byte, len(req.PlaintextKey))
	for i := range req.PlaintextKey {
		wrapped[i] = req.PlaintextKey[i] ^ wrapKey[i]
	}

	return KeyWrapResult{
		WrappedKey:         wrapped,
		Scheme:             KeyWrapSchemeEC,
		EphemeralPublicKey: ephemeralPubKey,
	}, nil
}

func (p *defaultCryptoProvider) wrapKeyWithRSA(req KeyWrapRequest) (KeyWrapResult, error) {
	encryptor, err := ocrypto.FromPublicPEM(req.PublicKeyPEM)
	if err != nil {
		return KeyWrapResult{}, fmt.Errorf("failed to create RSA encryptor: %w", err)
	}

	wrapped, err := encryptor.Encrypt(req.PlaintextKey)
	if err != nil {
		return KeyWrapResult{}, fmt.Errorf("failed to RSA encrypt key: %w", err)
	}

	return KeyWrapResult{
		WrappedKey: wrapped,
		Scheme:     KeyWrapSchemeRSA,
	}, nil
}
