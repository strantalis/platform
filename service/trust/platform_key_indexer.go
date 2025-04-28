package trust

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/opentdf/platform/protocol/go/policy"
	"github.com/opentdf/platform/protocol/go/policy/kasregistry"
	"github.com/opentdf/platform/sdk"
	"github.com/opentdf/platform/service/logger"
)

// Used for reaching out to platform to get keys

type PlatformKeyIndexer struct {
	// KeyIndex is the key index used to manage keys
	KeyIndex
	// SDK is the SDK instance used to interact with the platform
	sdk *sdk.SDK
	// Logger is the logger instance used for logging
	log *logger.Logger
}

// platformKeyAdapter is an adapter for KeyDetails, where keys come from the platform
type AsymKeyAdapter struct {
	asymKey *policy.AsymmetricKey
	log     *logger.Logger
}

// Used on decrypt of "this" KAS. Need the platform endpoint
// Indexer that can be instantiated with any crypto provider.\
// Common use is StandardCryptoProvider
func NewPlatformKeyIndexer(sdk *sdk.SDK, l *logger.Logger) *PlatformKeyIndexer {
	return &PlatformKeyIndexer{
		sdk: sdk,
		log: l,
	}
}

func (p *PlatformKeyIndexer) FindKeyByAlgorithm(ctx context.Context, algorithm string, includeLegacy bool) (KeyDetails, error) {
	return nil, errors.New("not implemented")
}

func (p *PlatformKeyIndexer) FindKeyByID(ctx context.Context, id KeyIdentifier) (KeyDetails, error) {
	req := &kasregistry.GetKeyRequest{
		Identifier: &kasregistry.GetKeyRequest_KeyId{
			KeyId: string(id),
		},
	}

	resp, err := p.sdk.KeyAccessServerRegistry.GetKey(ctx, req)
	if err != nil {
		return nil, err
	}

	return &AsymKeyAdapter{
		asymKey: resp.GetKey(),
		log:     p.log,
	}, nil
}

func (p *PlatformKeyIndexer) ListKeys(ctx context.Context) ([]KeyDetails, error) {
	return nil, errors.New("not implemented")
}

func (p *AsymKeyAdapter) ID() KeyIdentifier {
	return KeyIdentifier(p.asymKey.GetKeyId())
}
func (p *AsymKeyAdapter) Algorithm() string {
	return p.asymKey.GetKeyAlgorithm().String()
}
func (p *AsymKeyAdapter) IsLegacy() bool {
	return false
}

// This will point to the correct "manager"
func (p *AsymKeyAdapter) Mode() string {
	var mode string
	if p.asymKey.GetProviderConfig() != nil {
		mode = p.asymKey.GetProviderConfig().GetName()
	}
	return mode
}

func (p *AsymKeyAdapter) GetProviderConfig() *policy.KeyProviderConfig {
	return p.asymKey.GetProviderConfig()
}

func (p *AsymKeyAdapter) GetPrivateKeyCtx() ([]byte, error) {
	return p.asymKey.GetPrivateKeyCtx(), nil
}

// Needs to be unmarshalled.
// Probably should just be a base64 encoded string and not jsonb
// Need to handle remote / local crypto operations
// This assumes all local rn.
func (p *AsymKeyAdapter) ExportPublicKey(ctx context.Context, format KeyType) (string, error) {
	d := NewDefault(p.log)

	// Get public key.
	publicKeyCtx := p.asymKey.GetPublicKeyCtx()
	var pubKeyCtxMap map[string]any
	if err := json.Unmarshal(publicKeyCtx, &pubKeyCtxMap); err != nil {
		return "", err
	}

	pubKey, ok := pubKeyCtxMap["pubKey"].(string)
	if !ok {
		return "", errors.New("public key is not a string")
	}

	switch format {
	case KeyTypeJWK:
		// For JWK format (currently only supported for RSA)
		if p.asymKey.GetKeyAlgorithm() == policy.Algorithm_ALGORITHM_RSA_2048 {
			return d.RSAPublicKeyAsJSON(ctx, pubKey)
		}
		// For EC keys, we return the public key in PEM format
		jwkKey, err := convertPEMToJWK(pubKey)
		if err != nil {
			return "", err
		}

		return jwkKey, nil
	case KeyTypePKCS8:
		return pubKey, nil
	default:
		return "", errors.New("unsupported key type")
	}
}

func (p *AsymKeyAdapter) ExportCertificate(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}
