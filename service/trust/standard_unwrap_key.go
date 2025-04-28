package trust

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"

	"github.com/opentdf/platform/lib/ocrypto"
)

type StandardUnwrappedKey struct {
	rawKey []byte
	logger *slog.Logger
}

// NewStandardUnwrappedKey creates a new instance of StandardUnwrappedKey
func NewStandardUnwrappedKey(rawKey []byte) *StandardUnwrappedKey {
	return &StandardUnwrappedKey{
		rawKey: rawKey,
		logger: slog.Default(),
	}
}

func (k *StandardUnwrappedKey) DecryptAESGCM(iv []byte, body []byte, tagSize int) ([]byte, error) {
	aesGcm, err := ocrypto.NewAESGcm(k.rawKey)
	if err != nil {
		return nil, err
	}

	decryptedData, err := aesGcm.DecryptWithIVAndTagSize(iv, body, tagSize)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}

// Export returns the raw key data, optionally encrypting it with the provided encryptor
func (k *StandardUnwrappedKey) Export(encryptor Encapsulator) ([]byte, error) {
	if encryptor == nil {
		return k.rawKey, nil
	}

	// If an encryptor is provided, encrypt the key data before returning
	encryptedKey, err := encryptor.Encrypt(k.rawKey)
	if err != nil {
		if k.logger != nil {
			k.logger.Warn("failed to encrypt key data for export", "err", err)
		}
		return nil, err
	}

	return encryptedKey, nil
}

// VerifyBinding checks if the policy binding matches the given policy data
func (k *StandardUnwrappedKey) VerifyBinding(ctx context.Context, policy, policyBinding []byte) error {
	if len(k.rawKey) == 0 {
		return errors.New("key data is empty")
	}

	actualHMAC, err := k.generateHMACDigest(ctx, policy)
	if err != nil {
		return fmt.Errorf("unable to generate policy hmac: %w", err)
	}

	if !hmac.Equal(actualHMAC, policyBinding) {
		return errors.New("policy hmac mismatch")
	}

	return nil
}

// generateHMACDigest is a helper to generate an HMAC digest from a message using the key
func (k *StandardUnwrappedKey) generateHMACDigest(ctx context.Context, msg []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, k.rawKey)
	_, err := mac.Write(msg)
	if err != nil {
		if k.logger != nil {
			k.logger.WarnContext(ctx, "failed to compute hmac")
		}
		return nil, errors.New("policy hmac")
	}
	return mac.Sum(nil), nil
}
