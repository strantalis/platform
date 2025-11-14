package sdk

import (
	"context"

	exptdf "github.com/opentdf/platform/sdk/experimental/tdf"
)

// ExperimentalTDFWriter exposes the streaming TDF writer from the experimental
// package through the top-level SDK package. The implementation remains
// experimental and may change or be removed in future releases.
type ExperimentalTDFWriter = exptdf.Writer

// ExperimentalTDFWriterConfig aliases the experimental writer configuration.
type ExperimentalTDFWriterConfig = exptdf.WriterConfig

// ExperimentalTDFFinalizeConfig aliases the configuration used during
// writer finalization.
type ExperimentalTDFFinalizeConfig = exptdf.WriterFinalizeConfig

// ExperimentalTDFSegmentResult re-exports the per-segment result structure.
type ExperimentalTDFSegmentResult = exptdf.SegmentResult

// ExperimentalTDFFinalizeResult re-exports the finalization output structure.
type ExperimentalTDFFinalizeResult = exptdf.FinalizeResult

// ExperimentalTDFOption mirrors the functional option type used by the
// experimental writer APIs.
type ExperimentalTDFOption[T any] = exptdf.Option[T]

// ExperimentalTDFWriterOption is the functional option type for writer creation.
type ExperimentalTDFWriterOption = exptdf.Option[*exptdf.WriterConfig]

// ExperimentalTDFFinalizeOption is the functional option type for finalization.
type ExperimentalTDFFinalizeOption = exptdf.Option[*exptdf.WriterFinalizeConfig]

// ExperimentalTDFIntegrityAlgorithm mirrors the integrity algorithm enum.
type ExperimentalTDFIntegrityAlgorithm = exptdf.IntegrityAlgorithm

const (
	// ExperimentalIntegrityHS256 selects HMAC-SHA256 for integrity checks.
	ExperimentalIntegrityHS256 ExperimentalTDFIntegrityAlgorithm = exptdf.HS256
	// ExperimentalIntegrityGMAC selects GMAC for integrity checks.
	ExperimentalIntegrityGMAC ExperimentalTDFIntegrityAlgorithm = exptdf.GMAC
)

// NewExperimentalTDFWriter constructs a streaming TDF writer backed by the
// experimental implementation. Callers should treat the returned writer as
// experimental and subject to change in future releases.
func NewExperimentalTDFWriter(ctx context.Context, opts ...ExperimentalTDFWriterOption) (*ExperimentalTDFWriter, error) {
	return exptdf.NewWriter(ctx, opts...)
}

// ExperimentalWithIntegrityAlgorithm re-exports the integrity algorithm option.
var ExperimentalWithIntegrityAlgorithm = exptdf.WithIntegrityAlgorithm

// ExperimentalWithSegmentIntegrityAlgorithm re-exports the segment integrity option.
var ExperimentalWithSegmentIntegrityAlgorithm = exptdf.WithSegmentIntegrityAlgorithm

// ExperimentalWithInitialAttributes re-exports the initial attribute option.
var ExperimentalWithInitialAttributes = exptdf.WithInitialAttributes

// ExperimentalWithDefaultKASForWriter re-exports the default KAS writer option.
var ExperimentalWithDefaultKASForWriter = exptdf.WithDefaultKASForWriter

// ExperimentalWithCryptoProvider re-exports the crypto provider option.
var ExperimentalWithCryptoProvider = exptdf.WithCryptoProvider

// ExperimentalWithEntropySource re-exports the entropy source option.
var ExperimentalWithEntropySource = exptdf.WithEntropySource

// ExperimentalWithEncryptedMetadata re-exports the encrypted metadata finalization option.
var ExperimentalWithEncryptedMetadata = exptdf.WithEncryptedMetadata

// ExperimentalWithPayloadMimeType re-exports the payload MIME type option.
var ExperimentalWithPayloadMimeType = exptdf.WithPayloadMimeType

// ExperimentalWithSegments re-exports the segment selection option.
var ExperimentalWithSegments = exptdf.WithSegments

// ExperimentalWithDefaultKAS re-exports the default KAS finalization option.
var ExperimentalWithDefaultKAS = exptdf.WithDefaultKAS

// ExperimentalWithAttributeValues re-exports the attribute configuration option.
var ExperimentalWithAttributeValues = exptdf.WithAttributeValues

// ExperimentalWithExcludeVersionFromManifest re-exports the manifest version option.
var ExperimentalWithExcludeVersionFromManifest = exptdf.WithExcludeVersionFromManifest

// ExperimentalWithAssertions re-exports the assertion configuration option.
var ExperimentalWithAssertions = exptdf.WithAssertions

// ExperimentalCryptoProvider aliases the experimental crypto provider interface.
type ExperimentalCryptoProvider = exptdf.CryptoProvider

// ExperimentalEntropySource aliases the entropy source interface.
type ExperimentalEntropySource = exptdf.EntropySource

// ExperimentalAEADCipherFactory aliases the AES-GCM factory interface.
type ExperimentalAEADCipherFactory = exptdf.AEADCipherFactory

// ExperimentalAESGCM aliases the AES-GCM cipher interface.
type ExperimentalAESGCM = exptdf.AESGCM

// ExperimentalEncodingProvider aliases the encoding provider interface.
type ExperimentalEncodingProvider = exptdf.EncodingProvider

// ExperimentalIntegrityProvider aliases the integrity provider interface.
type ExperimentalIntegrityProvider = exptdf.IntegrityProvider

// ExperimentalKeyWrapProvider aliases the key wrap provider interface.
type ExperimentalKeyWrapProvider = exptdf.KeyWrapProvider

// ExperimentalKeyWrapRequest re-exports the key wrap request structure.
type ExperimentalKeyWrapRequest = exptdf.KeyWrapRequest

// ExperimentalKeyWrapResult re-exports the key wrap result structure.
type ExperimentalKeyWrapResult = exptdf.KeyWrapResult

// ExperimentalKeyWrapScheme aliases the key wrap scheme enumeration.
type ExperimentalKeyWrapScheme = exptdf.KeyWrapScheme

const (
	// ExperimentalKeyWrapSchemeRSA indicates RSA-based key wrapping.
	ExperimentalKeyWrapSchemeRSA ExperimentalKeyWrapScheme = exptdf.KeyWrapSchemeRSA
	// ExperimentalKeyWrapSchemeEC indicates EC-based key wrapping.
	ExperimentalKeyWrapSchemeEC ExperimentalKeyWrapScheme = exptdf.KeyWrapSchemeEC
)

// ExperimentalAssertionConfig aliases the assertion configuration type.
type ExperimentalAssertionConfig = exptdf.AssertionConfig

// ExperimentalAssertion aliases the assertion manifest structure.
type ExperimentalAssertion = exptdf.Assertion

// ExperimentalAssertionType aliases the assertion type enumeration.
type ExperimentalAssertionType = exptdf.AssertionType

const (
	// ExperimentalHandlingAssertion mirrors the handling assertion type.
	ExperimentalHandlingAssertion ExperimentalAssertionType = exptdf.HandlingAssertion
	// ExperimentalBaseAssertion mirrors the base assertion type.
	ExperimentalBaseAssertion ExperimentalAssertionType = exptdf.BaseAssertion
)

// ExperimentalScope aliases the assertion scope enumeration.
type ExperimentalScope = exptdf.Scope

const (
	// ExperimentalTrustedDataObjScope mirrors the trusted data object scope.
	ExperimentalTrustedDataObjScope ExperimentalScope = exptdf.TrustedDataObjScope
	// ExperimentalPayloadScope mirrors the payload scope.
	ExperimentalPayloadScope ExperimentalScope = exptdf.PayloadScope
)

// ExperimentalAppliesToState aliases the assertion lifecycle enumeration.
type ExperimentalAppliesToState = exptdf.AppliesToState

const (
	// ExperimentalEncrypted mirrors the encrypted lifecycle state.
	ExperimentalEncrypted ExperimentalAppliesToState = exptdf.Encrypted
	// ExperimentalUnencrypted mirrors the unencrypted lifecycle state.
	ExperimentalUnencrypted ExperimentalAppliesToState = exptdf.Unencrypted
)

// ExperimentalBinding aliases the assertion binding structure.
type ExperimentalBinding = exptdf.Binding

// ExperimentalStatement aliases the assertion statement structure.
type ExperimentalStatement = exptdf.Statement

// ExperimentalBindingMethod aliases the binding method enumeration.
type ExperimentalBindingMethod = exptdf.BindingMethod

const (
	// ExperimentalBindingMethodJWS mirrors the JWS binding method.
	ExperimentalBindingMethodJWS ExperimentalBindingMethod = exptdf.JWS
)

// ExperimentalAssertionKey aliases the assertion key structure.
type ExperimentalAssertionKey = exptdf.AssertionKey

// ExperimentalAssertionKeyAlg aliases the assertion key algorithm enumeration.
type ExperimentalAssertionKeyAlg = exptdf.AssertionKeyAlg

const (
	// ExperimentalAssertionKeyAlgRS256 mirrors the RS256 assertion key algorithm.
	ExperimentalAssertionKeyAlgRS256 ExperimentalAssertionKeyAlg = exptdf.AssertionKeyAlgRS256
	// ExperimentalAssertionKeyAlgHS256 mirrors the HS256 assertion key algorithm.
	ExperimentalAssertionKeyAlgHS256 ExperimentalAssertionKeyAlg = exptdf.AssertionKeyAlgHS256
)

// ExperimentalAssertionVerificationKeys aliases the verification key set structure.
type ExperimentalAssertionVerificationKeys = exptdf.AssertionVerificationKeys

// ExperimentalManifest aliases the TDF manifest structure.
type ExperimentalManifest = exptdf.Manifest

// ExperimentalKeyAccess aliases the key access manifest structure.
type ExperimentalKeyAccess = exptdf.KeyAccess

// ExperimentalRootSignature aliases the root signature structure.
type ExperimentalRootSignature = exptdf.RootSignature

// ExperimentalIntegrityInformation aliases the integrity info structure.
type ExperimentalIntegrityInformation = exptdf.IntegrityInformation

// ExperimentalEncryptionInformation aliases the encryption info structure.
type ExperimentalEncryptionInformation = exptdf.EncryptionInformation

// ExperimentalPayload aliases the payload manifest structure.
type ExperimentalPayload = exptdf.Payload

// ExperimentalSegment aliases the segment manifest structure.
type ExperimentalSegment = exptdf.Segment
