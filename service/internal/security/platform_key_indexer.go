package security

import (
	"context"
	"errors"

	"github.com/opentdf/platform/protocol/go/policy"
	"github.com/opentdf/platform/protocol/go/policy/kasregistry"
	"github.com/opentdf/platform/sdk"
	"github.com/opentdf/platform/service/trust"
)

// Used for reaching out to platform to get keys

type PlatformKeyIndexer struct {
	// KeyIndex is the key index used to manage keys
	trust.KeyIndex
	// SDK is the SDK instance used to interact with the platform
	sdk *sdk.SDK
}

// platformKeyAdapter is an adapter for KeyDetails, where keys come from the platform
type AsymKeyAdapter struct {
	asymKey *policy.AsymmetricKey
}

// Used on decrypt of "this" KAS. Need the platform endpoint
// Indexer that can be instantiated with any crypto provider.\
// Common use is StandardCryptoProvider
func NewPlatformKeyIndexer(sdk *sdk.SDK) *PlatformKeyIndexer {
	return &PlatformKeyIndexer{
		sdk: sdk,
	}
}

func (p *PlatformKeyIndexer) FindKeyByAlgorithm(ctx context.Context, algorithm string, includeLegacy bool) (trust.KeyDetails, error) {
	return nil, errors.New("not implemented")
}

func (p *PlatformKeyIndexer) FindKeyByID(ctx context.Context, id trust.KeyIdentifier) (trust.KeyDetails, error) {
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
	}, nil
}

func (p *PlatformKeyIndexer) ListKeys(ctx context.Context) ([]trust.KeyDetails, error) {
	return nil, errors.New("not implemented")
}

func (p *AsymKeyAdapter) ID() trust.KeyIdentifier {
	return trust.KeyIdentifier(p.asymKey.GetKeyId())
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
func (p *AsymKeyAdapter) ExportPublicKey(ctx context.Context, format trust.KeyType) (string, error) {
	return string(p.asymKey.GetPublicKeyCtx()), nil
}
func (p *AsymKeyAdapter) ExportCertificate(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}
