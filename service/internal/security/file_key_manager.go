package security

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"

	"github.com/opentdf/platform/sdk"
	"github.com/opentdf/platform/service/trust"
)

type FileKeyManager struct {
	*PlatformKeyIndexer
	log *slog.Logger
	*Default
}

const (
	fileKeyManagerName = "opentdf.io/file"
)

func NewFileKeyManager(sdk *sdk.SDK, l *slog.Logger) *FileKeyManager {
	return &FileKeyManager{
		PlatformKeyIndexer: NewPlatformKeyIndexer(sdk),
		log:                l,
		Default:            NewDefault(l),
	}
}

func (m *FileKeyManager) Name() string {
	return fileKeyManagerName
}

func (m *FileKeyManager) Decrypt(ctx context.Context, keyID trust.KeyIdentifier, ciphertext []byte, ephemeralPublicKey []byte) (trust.ProtectedKey, error) {
	kid := string(keyID)

	// Get key.
	keyDetails, err := m.FindKeyByID(ctx, trust.KeyIdentifier(kid))
	if err != nil {
		return nil, err
	}

	// Load sym key from file.
	// Cast key details to an AsymKeyAdapter object.
	asymKeyAdapter, ok := keyDetails.(*AsymKeyAdapter)
	if !ok {
		return nil, errors.New("failed to cast key details to AsymKeyAdapter")
	}

	// Get provider config and unmarshal to a map.
	providerConfig := asymKeyAdapter.GetProviderConfig()
	var configMap map[string]any
	if err := json.Unmarshal([]byte(providerConfig.GetConfigJson()), &configMap); err != nil {
		return nil, errors.New("failed to unmarshal provider config to map")
	}

	// Look for the "filepath" key.
	filepath, ok := configMap["filepath"].(string)
	if !ok || filepath == "" {
		return nil, errors.New("filepath not found or invalid in provider config")
	}

	// Load the bytes from the specified filepath.
	fileBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, errors.New("failed to read file from filepath: " + err.Error())
	}

	// Now decrypt the key.
	// Get the private key context from the AsymKeyAdapter.
	privateKeyContext, err := asymKeyAdapter.GetPrivateKeyCtx()
	if err != nil {
		return nil, errors.New("failed to get private key context")
	}

	// Unmarshal the private key context to a map.
	var privateKeyMap map[string]any
	if err := json.Unmarshal([]byte(privateKeyContext), &privateKeyMap); err != nil {
		return nil, errors.New("failed to unmarshal private key context to map")
	}
	// Look for the "wrappedKey" key.
	wrappedKey, ok := privateKeyMap["wrappedKey"].(string)
	if !ok || wrappedKey == "" {
		return nil, errors.New("wrappedKey not found or invalid in private key context")
	}

	// Decrypt the wrappedKey using the key found in fileBytes.
	unwrappedAsymKey, err := m.DecryptSymmetric(ctx, fileBytes, []byte(wrappedKey))
	if err != nil {
		return nil, errors.New("failed to decrypt wrappedKey: " + err.Error())
	}

	unwrappedDek, err := m.DecryptAsymmetric(ctx, asymKeyAdapter.Algorithm(), unwrappedAsymKey, ciphertext, ephemeralPublicKey)
	if err != nil {
		return nil, errors.New("failed to decrypt asymmetric key: " + err.Error())
	}

	return NewStandardUnwrappedKey(unwrappedDek), nil
}
