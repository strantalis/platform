package access

import (
	"context"
	"net/url"

	kaspb "github.com/opentdf/platform/protocol/go/kas"
	otdf "github.com/opentdf/platform/sdk"
	"github.com/opentdf/platform/service/internal/security"
	"github.com/opentdf/platform/service/logger"
	"github.com/opentdf/platform/service/pkg/config"
	"github.com/opentdf/platform/service/trust"
	"go.opentelemetry.io/otel/trace"
)

const (
	ErrHSM    = Error("hsm unexpected")
	ErrConfig = Error("invalid config")
)

type Provider struct {
	kaspb.AccessServiceServer
	URI          url.URL `json:"uri"`
	SDK          *otdf.SDK
	AttributeSvc *url.URL
	KeyIndex     trust.KeyIndex
	KeyManager   trust.KeyManager
	// Deprecated: Use SecurityProvider instead
	CryptoProvider *security.StandardCrypto // Kept for backward compatibility
	Logger         *logger.Logger
	Config         *config.ServiceConfig
	KASConfig
	trace.Tracer
}

// GetSecurityProvider returns the SecurityProvider
func (p *Provider) GetSecurityProvider() trust.KeyManager {
	p.initSecurityProviderAdapter()
	return p.KeyManager
}

func (p *Provider) GetKeyIndex() trust.KeyIndex {
	p.initSecurityProviderAdapter()
	return p.KeyIndex
}

type KASConfig struct {
	// Which keys are currently the default.
	Keyring []CurrentKeyFor `mapstructure:"keyring" json:"keyring"`
	// Deprecated
	ECCertID string `mapstructure:"eccertid" json:"eccertid"`
	// Deprecated
	RSACertID string `mapstructure:"rsacertid" json:"rsacertid"`

	// Enables experimental EC rewrap support in TDFs
	// Enabling is required to parse KAOs with the `ec-wrapped` type,
	// and (currently) also enables responding with ECIES encrypted responses.
	ECTDFEnabled bool `mapstructure:"ec_tdf_enabled" json:"ec_tdf_enabled"`
}

// Specifies the preferred/default key for a given algorithm type.
type CurrentKeyFor struct {
	Algorithm string `mapstructure:"alg" json:"alg"`
	KID       string `mapstructure:"kid" json:"kid"`
	// Indicates that the key should not be serves by default,
	// but instead is allowed for legacy reasons on decrypt (rewrap) only
	Legacy bool `mapstructure:"legacy" json:"legacy"`
}

func (p *Provider) IsReady(ctx context.Context) error {
	// TODO: Not sure what we want to check here?
	p.Logger.TraceContext(ctx, "checking readiness of kas service")
	return nil
}

func (kasCfg *KASConfig) UpgradeMapToKeyring(c *security.StandardCrypto) {
	switch {
	case kasCfg.ECCertID != "" && len(kasCfg.Keyring) > 0:
		panic("invalid kas cfg: please specify keyring or eccertid, not both")
	case len(kasCfg.Keyring) == 0:
		deprecatedOrDefault := func(kid, alg string) {
			if kid == "" {
				kid = c.FindKID(alg)
			}
			if kid == "" {
				// no known key for this algorithm type
				return
			}
			kasCfg.Keyring = append(kasCfg.Keyring, CurrentKeyFor{
				Algorithm: alg,
				KID:       kid,
			})
			kasCfg.Keyring = append(kasCfg.Keyring, CurrentKeyFor{
				Algorithm: alg,
				KID:       kid,
				Legacy:    true,
			})
		}
		deprecatedOrDefault(kasCfg.ECCertID, security.AlgorithmECP256R1)
		deprecatedOrDefault(kasCfg.RSACertID, security.AlgorithmRSA2048)
	default:
		kasCfg.Keyring = append(kasCfg.Keyring, inferLegacyKeys(kasCfg.Keyring)...)
	}
}

func (p *Provider) initSecurityProviderAdapter() {
	// If the CryptoProvider is set, create a SecurityProviderAdapter
	if p.CryptoProvider == nil || p.KeyManager != nil && p.KeyIndex != nil {
		return
	}
	var defaults []string
	var legacies []string
	for _, key := range p.KASConfig.Keyring {
		if key.Legacy {
			legacies = append(legacies, key.KID)
		} else {
			defaults = append(defaults, key.KID)
		}
	}
	if len(defaults) == 0 && len(legacies) == 0 {
		for _, alg := range []string{security.AlgorithmECP256R1, security.AlgorithmRSA2048} {
			kid := p.CryptoProvider.FindKID(alg)
			if kid != "" {
				defaults = append(defaults, kid)
			} else {
				p.Logger.Warn("no default key found for algorithm", "algorithm", alg)
			}
		}
	}

	inProcessService := security.NewSecurityProviderAdapter(p.CryptoProvider, defaults, legacies)

	if p.KeyIndex == nil {
		p.Logger.Warn("fallback to in-process key index")
		p.KeyIndex = inProcessService
	}
	if p.KeyManager == nil {
		p.Logger.Error("fallback to in-process manager")
		p.KeyManager = inProcessService
	}
}

// If there exists *any* legacy keys, returns empty list.
// Otherwise, create a copy with legacy=true for all values
func inferLegacyKeys(keys []CurrentKeyFor) []CurrentKeyFor {
	for _, k := range keys {
		if k.Legacy {
			return nil
		}
	}
	l := make([]CurrentKeyFor, len(keys))
	for i, k := range keys {
		l[i] = k
		l[i].Legacy = true
	}
	return l
}
