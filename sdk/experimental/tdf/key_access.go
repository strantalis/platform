// Experimental: This package is EXPERIMENTAL and may change or be removed at any time

package tdf

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/opentdf/platform/sdk/experimental/tdf/keysplit"
)

var tdfSaltBytes []byte

// tdfSalt generates the standard TDF salt for key derivation
func init() {
	digest := sha256.New()
	digest.Write([]byte("TDF"))
	tdfSaltBytes = digest.Sum(nil)
}

func tdfSalt() []byte {
	return tdfSaltBytes
}

// BuildKeyAccessObjects creates KeyAccess objects from splits for TDF manifest inclusion
func buildKeyAccessObjects(provider CryptoProvider, result *keysplit.SplitResult, policyBytes []byte, metadata string) ([]KeyAccess, error) {
	if result == nil || len(result.Splits) == 0 {
		return nil, errors.New("no splits provided")
	}

	var keyAccessList []KeyAccess

	// Create base64-encoded policy for binding
	base64Policy := string(provider.Base64Encode(policyBytes))

	for _, split := range result.Splits {
		for _, kasURL := range split.KASURLs {
			// Get public key info for this KAS
			pubKeyInfo, exists := result.KASPublicKeys[kasURL]
			if !exists {
				slog.Warn("no public key found for KAS, skipping",
					slog.String("kas_url", kasURL),
					slog.String("split_id", split.ID))
				continue
			}

			// Create policy binding
			policyBinding, err := createPolicyBinding(provider, split.Data, base64Policy)
			if err != nil {
				return nil, fmt.Errorf("failed to create policy binding for KAS %s: %w", kasURL, err)
			}

			// Encrypt metadata if provided
			var encryptedMetadata string
			if metadata != "" {
				var err error
				encryptedMetadata, err = encryptMetadata(provider, split.Data, metadata)
				if err != nil {
					return nil, fmt.Errorf("failed to encrypt metadata for KAS %s: %w", kasURL, err)
				}
			}

			// Encrypt the split key with KAS public key
			wrappedKey, keyType, ephemeralPubKey, err := wrapKeyWithPublicKey(provider, split.Data, pubKeyInfo)
			if err != nil {
				return nil, fmt.Errorf("failed to wrap key for KAS %s: %w", kasURL, err)
			}

			// Build the KeyAccess object
			keyAccess := KeyAccess{
				KeyType:           keyType,
				KasURL:            kasURL,
				KID:               pubKeyInfo.KID,
				Protocol:          "kas",
				SplitID:           split.ID,
				WrappedKey:        wrappedKey,
				PolicyBinding:     policyBinding,
				EncryptedMetadata: encryptedMetadata,
			}

			// Add ephemeral public key for EC keys
			if ephemeralPubKey != "" {
				keyAccess.EphemeralPublicKey = ephemeralPubKey
			}

			keyAccessList = append(keyAccessList, keyAccess)

			slog.Debug("created key access object",
				slog.String("kas_url", kasURL),
				slog.String("split_id", split.ID),
				slog.String("key_type", keyType),
				slog.String("kid", pubKeyInfo.KID))
		}
	}

	if len(keyAccessList) == 0 {
		return nil, errors.New("no valid key access objects generated")
	}

	slog.Debug("built key access objects",
		slog.Int("num_key_access", len(keyAccessList)),
		slog.Int("num_splits", len(result.Splits)))

	return keyAccessList, nil
}

// createPolicyBinding creates an HMAC binding between the key and policy
func createPolicyBinding(provider CryptoProvider, symKey []byte, base64PolicyObject string) (any, error) {
	// Create HMAC hash of the policy using the symmetric key
	hmacHash, err := provider.HMACSHA256(symKey, []byte(base64PolicyObject))
	if err != nil {
		return nil, fmt.Errorf("failed to compute policy binding hmac: %w", err)
	}

	// Convert to hex string
	hashHex := hex.EncodeToString(hmacHash)

	// Create policy binding structure
	binding := PolicyBinding{
		Alg:  kPolicyBindingAlg,
		Hash: string(provider.Base64Encode([]byte(hashHex))),
	}

	return binding, nil
}

// encryptMetadata encrypts TDF metadata using the split key
func encryptMetadata(provider CryptoProvider, symKey []byte, metadata string) (string, error) {
	// Create AES-GCM cipher
	gcm, err := provider.NewAESGCM(symKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES-GCM: %w", err)
	}

	// Encrypt the metadata
	encryptedBytes, err := gcm.Encrypt([]byte(metadata))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt metadata: %w", err)
	}

	// Extract IV (first 12 bytes for GCM)
	const gcmStandardNonceSize = 12
	if len(encryptedBytes) < gcmStandardNonceSize {
		return "", errors.New("encrypted metadata too short for nonce extraction")
	}
	iv := encryptedBytes[:gcmStandardNonceSize]

	// Create encrypted metadata structure
	encMeta := EncryptedMetadata{
		Cipher: string(provider.Base64Encode(encryptedBytes)),
		Iv:     string(provider.Base64Encode(iv)),
	}

	// Serialize to JSON and base64 encode
	metadataJSON, err := json.Marshal(encMeta)
	if err != nil {
		return "", fmt.Errorf("failed to marshal encrypted metadata: %w", err)
	}

	return string(provider.Base64Encode(metadataJSON)), nil
}

// wrapKeyWithPublicKey encrypts a symmetric key with a KAS public key
func wrapKeyWithPublicKey(provider CryptoProvider, symKey []byte, pubKeyInfo keysplit.KASPublicKey) (string, string, string, error) {
	if pubKeyInfo.PEM == "" {
		return "", "", "", fmt.Errorf("public key PEM is empty for KAS %s", pubKeyInfo.URL)
	}

	req := KeyWrapRequest{
		Algorithm:    pubKeyInfo.Algorithm,
		PublicKeyPEM: pubKeyInfo.PEM,
		PlaintextKey: symKey,
		Salt:         tdfSalt(),
	}

	res, err := provider.WrapKey(req)
	if err != nil {
		return "", "", "", err
	}

	wrapped := provider.Base64Encode(res.WrappedKey)
	return string(wrapped), string(res.Scheme), res.EphemeralPublicKey, nil
}
