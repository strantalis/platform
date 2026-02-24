package dev

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/opentdf/platform/lib/ocrypto"
	"github.com/opentdf/platform/protocol/go/policy"
	"github.com/opentdf/platform/protocol/go/policy/kasregistry"
	policyunsafe "github.com/opentdf/platform/protocol/go/policy/unsafe"
	"github.com/opentdf/platform/sdk"
)

const (
	defaultRootKeyID = "dev-root-key"
	defaultKasName   = "dev-kas"
)

type KasSeedOptions struct {
	PlatformEndpoint string
	KasURI           string
	ClientID         string
	ClientSecret     string
	RootKey          string
	Regenerate       bool
}

type KasSeedKey struct {
	Kid       string
	Algorithm policy.Algorithm
	Created   bool
}

type KasSeedResult struct {
	KasID  string
	KasURI string
	Keys   []KasSeedKey
}

func EnsureKasRegistryKeys(ctx context.Context, opts KasSeedOptions) (*KasSeedResult, error) {
	if opts.PlatformEndpoint == "" {
		return nil, errors.New("platform endpoint is required")
	}
	if opts.KasURI == "" {
		return nil, errors.New("kas uri is required")
	}
	if opts.ClientID == "" || opts.ClientSecret == "" {
		return nil, errors.New("client credentials are required")
	}

	rootKey := strings.TrimSpace(opts.RootKey)
	rootKeyBytes, err := hex.DecodeString(rootKey)
	if err != nil {
		return nil, fmt.Errorf("decode root key: %w", err)
	}
	if len(rootKeyBytes) != defaultRootKeyBytes {
		return nil, fmt.Errorf("root key must be 32 bytes, got %d", len(rootKeyBytes))
	}

	//nolint:contextcheck // sdk.New does not accept a context and validates connectivity internally.
	s, err := sdk.New(
		opts.PlatformEndpoint,
		sdk.WithClientCredentials(opts.ClientID, opts.ClientSecret, nil),
		sdk.WithInsecurePlaintextConn(),
	)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	kas, err := ensureKasRegistry(ctx, s, opts.KasURI)
	if err != nil {
		return nil, err
	}

	keysToEnsure := []struct {
		kid string
		alg policy.Algorithm
	}{
		{kid: defaultKASRSAKID, alg: policy.Algorithm_ALGORITHM_RSA_2048},
		{kid: defaultKASECKID, alg: policy.Algorithm_ALGORITHM_EC_P256},
	}

	result := &KasSeedResult{
		KasID:  kas.GetId(),
		KasURI: opts.KasURI,
		Keys:   make([]KasSeedKey, 0, len(keysToEnsure)),
	}

	for _, spec := range keysToEnsure {
		created, err := ensureKasKey(ctx, s, kas.GetId(), opts.KasURI, rootKeyBytes, spec.kid, spec.alg, opts.Regenerate)
		if err != nil {
			return nil, err
		}
		result.Keys = append(result.Keys, KasSeedKey{
			Kid:       spec.kid,
			Algorithm: spec.alg,
			Created:   created,
		})
	}

	return result, nil
}

func ensureKasRegistry(ctx context.Context, s *sdk.SDK, kasURI string) (*policy.KeyAccessServer, error) {
	resp, err := s.KeyAccessServerRegistry.GetKeyAccessServer(ctx, &kasregistry.GetKeyAccessServerRequest{
		Identifier: &kasregistry.GetKeyAccessServerRequest_Uri{Uri: kasURI},
	})
	if err == nil {
		return resp.GetKeyAccessServer(), nil
	}
	if connect.CodeOf(err) != connect.CodeNotFound {
		return nil, err
	}

	created, err := s.KeyAccessServerRegistry.CreateKeyAccessServer(ctx, &kasregistry.CreateKeyAccessServerRequest{
		Uri:  kasURI,
		Name: defaultKasName,
	})
	if err != nil {
		return nil, err
	}
	return created.GetKeyAccessServer(), nil
}

func ensureKasKey(ctx context.Context, s *sdk.SDK, kasID, kasURI string, rootKey []byte, kid string, alg policy.Algorithm, regenerate bool) (bool, error) {
	existing, err := getKasKey(ctx, s, kasURI, kid)
	switch {
	case err == nil && !regenerate:
		return false, nil
	case err == nil && regenerate:
		if err := unsafeDeleteKasKey(ctx, s, existing, kasURI); err != nil {
			return false, err
		}
	case connect.CodeOf(err) != connect.CodeNotFound:
		return false, err
	}

	privatePEM, publicPEM, err := generateKeyPairPEM(alg)
	if err != nil {
		return false, err
	}
	wrappedKey, err := wrapPrivateKey(privatePEM, rootKey)
	if err != nil {
		return false, err
	}

	_, err = s.KeyAccessServerRegistry.CreateKey(ctx, &kasregistry.CreateKeyRequest{
		KasId:        kasID,
		KeyId:        kid,
		KeyAlgorithm: alg,
		KeyMode:      policy.KeyMode_KEY_MODE_CONFIG_ROOT_KEY,
		PublicKeyCtx: &policy.PublicKeyCtx{
			Pem: base64.StdEncoding.EncodeToString([]byte(publicPEM)),
		},
		PrivateKeyCtx: &policy.PrivateKeyCtx{
			KeyId:      defaultRootKeyID,
			WrappedKey: wrappedKey,
		},
		Legacy: false,
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func getKasKey(ctx context.Context, s *sdk.SDK, kasURI, kid string) (*policy.KasKey, error) {
	resp, err := s.KeyAccessServerRegistry.GetKey(ctx, &kasregistry.GetKeyRequest{
		Identifier: &kasregistry.GetKeyRequest_Key{
			Key: &kasregistry.KasKeyIdentifier{
				Identifier: &kasregistry.KasKeyIdentifier_Uri{Uri: kasURI},
				Kid:        kid,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetKasKey(), nil
}

func unsafeDeleteKasKey(ctx context.Context, s *sdk.SDK, key *policy.KasKey, kasURI string) error {
	if key == nil || key.GetKey() == nil {
		return errors.New("missing key details")
	}
	_, err := s.Unsafe.UnsafeDeleteKasKey(ctx, &policyunsafe.UnsafeDeleteKasKeyRequest{
		Id:     key.GetKey().GetId(),
		Kid:    key.GetKey().GetKeyId(),
		KasUri: kasURI,
	})
	return err
}

func generateKeyPairPEM(alg policy.Algorithm) (string, string, error) {
	switch alg {
	case policy.Algorithm_ALGORITHM_UNSPECIFIED:
		return "", "", fmt.Errorf("unsupported key algorithm: %s", alg)
	case policy.Algorithm_ALGORITHM_RSA_2048:
		keyPair, err := ocrypto.NewRSAKeyPair(defaultRSAKeySize)
		if err != nil {
			return "", "", fmt.Errorf("generate rsa key: %w", err)
		}
		privatePEM, err := keyPair.PrivateKeyInPemFormat()
		if err != nil {
			return "", "", fmt.Errorf("encode rsa private key: %w", err)
		}
		publicPEM, err := keyPair.PublicKeyInPemFormat()
		if err != nil {
			return "", "", fmt.Errorf("encode rsa public key: %w", err)
		}
		return privatePEM, publicPEM, nil
	case policy.Algorithm_ALGORITHM_RSA_4096:
		return "", "", fmt.Errorf("unsupported key algorithm: %s", alg)
	case policy.Algorithm_ALGORITHM_EC_P256:
		keyPair, err := ocrypto.NewECKeyPair(ocrypto.ECCModeSecp256r1)
		if err != nil {
			return "", "", fmt.Errorf("generate ec key: %w", err)
		}
		privatePEM, err := keyPair.PrivateKeyInPemFormat()
		if err != nil {
			return "", "", fmt.Errorf("encode ec private key: %w", err)
		}
		publicPEM, err := keyPair.PublicKeyInPemFormat()
		if err != nil {
			return "", "", fmt.Errorf("encode ec public key: %w", err)
		}
		return privatePEM, publicPEM, nil
	case policy.Algorithm_ALGORITHM_EC_P384:
		return "", "", fmt.Errorf("unsupported key algorithm: %s", alg)
	case policy.Algorithm_ALGORITHM_EC_P521:
		return "", "", fmt.Errorf("unsupported key algorithm: %s", alg)
	default:
		return "", "", fmt.Errorf("unsupported key algorithm: %s", alg)
	}
}

func wrapPrivateKey(privatePEM string, rootKey []byte) (string, error) {
	aesKey, err := ocrypto.NewAESGcm(rootKey)
	if err != nil {
		return "", fmt.Errorf("create aes-gcm key: %w", err)
	}
	wrapped, err := aesKey.Encrypt([]byte(privatePEM))
	if err != nil {
		return "", fmt.Errorf("wrap private key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(wrapped), nil
}
