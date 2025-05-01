package trust

import (
	"context"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	openbao "github.com/openbao/openbao/api/v2"
	"github.com/opentdf/platform/lib/ocrypto"
	"github.com/opentdf/platform/protocol/go/policy"
	"github.com/opentdf/platform/service/logger"
)

type BaoKey struct {
	WrappedKey string `json:"wrappedKey"`
	KeyID      string `json:"keyID"`
}

type BaoManager struct {
	client *openbao.Client
	*PlatformKeyIndexer
	l *logger.Logger
}

func NewBaoManager(index *PlatformKeyIndexer, l *logger.Logger) *BaoManager {
	config := openbao.DefaultConfig()

	config.Address = "http://127.0.0.1:8200"

	client, err := openbao.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize OpenBao client: %v", err)
	}

	client.SetToken("dev-only-token")

	return &BaoManager{
		client:             client,
		PlatformKeyIndexer: index,
		l:                  l,
	}
}

// Name is a unique identifier for the key manager.
// This can be used by the KeyDetail.Mode() method to determine which KeyManager to use,
// when multiple KeyManagers are installed.
func (b *BaoManager) Name() string {
	return "openbao"
}

// Decrypt decrypts data that was encrypted with the key identified by keyID
// For EC keys, ephemeralPublicKey must be non-nil
// For RSA keys, ephemeralPublicKey should be nil
// Returns an UnwrappedKeyData interface for further operations
func (b *BaoManager) Decrypt(ctx context.Context, keyID KeyIdentifier, ciphertext []byte, ephemeralPublicKey []byte) (ProtectedKey, error) {
	kid := string(keyID)

	// Get key.
	keyDetails, err := b.FindKeyByID(context.Background(), KeyIdentifier(kid))
	if err != nil {
		return nil, err
	}

	asymKey := keyDetails.(*AsymKeyAdapter)

	baoKey := &BaoKey{}

	bb, err := asymKey.GetPrivateKeyCtx()
	if err != nil {
		return nil, err
	}

	fmt.Println(string(bb))

	err = json.Unmarshal(bb, baoKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BaoKey: %w", err)
	}

	switch asymKey.KeyMode() {
	case policy.KeyMode_KEY_MODE_LOCAL:
		if asymKey.GetProviderConfig() != nil {
			wk, err := base64.StdEncoding.DecodeString(baoKey.WrappedKey[:])
			if err != nil {
				return nil, fmt.Errorf("failed to decode wrapped key: %w", err)
			}
			sec, err := b.client.Logical().Write(fmt.Sprintf("transit/decrypt/%s", baoKey.KeyID), map[string]interface{}{
				"ciphertext": fmt.Sprintf("%s", wk),
			})
			if err != nil {
				return nil, err
			}
			fmt.Println(sec.Data["plaintext"])
			plainTextKey, err := base64.StdEncoding.DecodeString(string(sec.Data["plaintext"].(string)))
			if err != nil {
				return nil, err
			}
			decryptor, err := ocrypto.FromPrivatePEM(string(plainTextKey))
			if err != nil {
				b.l.Error(fmt.Sprintf("failed to parse private key PEM: %v", err))
				return nil, err
			}
			plainText, err := decryptor.Decrypt(ciphertext)
			if err != nil {
				return nil, err
			}
			return NewStandardUnwrappedKey(plainText), nil
		}
	}

	return nil, fmt.Errorf("Decrypt not implemented")
}

// DeriveKey computes an agreed upon secret key, which NanoTDF may directly as the DEK or a key split
func (b *BaoManager) DeriveKey(ctx context.Context, kasKID KeyIdentifier, ephemeralPublicKeyBytes []byte, curve elliptic.Curve) (ProtectedKey, error) {
	return nil, fmt.Errorf("DeriveKey not implemented")
}

// GenerateECSessionKey generates a private session key, for use with a client-provided ephemeral public key
func (b *BaoManager) GenerateECSessionKey(ctx context.Context, ephemeralPublicKey string) (Encapsulator, error) {
	return nil, fmt.Errorf("GenerateECSessionKey not implemented")
}

// Close releases any resources held by the provider
func (b *BaoManager) Close() {}
