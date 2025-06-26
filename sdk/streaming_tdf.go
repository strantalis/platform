package sdk

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/opentdf/platform/lib/ocrypto"
	"github.com/opentdf/platform/sdk/internal/archive"
)

// StreamingTDFEncryptor handles the encryption of data in a streaming fashion,
// allowing for chunks of data to be encrypted and assembled into a TDF without
// knowing the total size in advance.
type StreamingTDFEncryptor struct {
	sdk        *SDK // Reference to the parent SDK instance
	writer     io.Writer
	tdfWriter  *archive.TDFWriter
	tdfConfig  *TDFConfig
	payloadKey [kKeySize]byte
	aesGcm     ocrypto.AesGcm
	manifest   Manifest

	// collectedSegments stores information about each encrypted chunk.
	collectedSegments []Segment
	// aggregateSegmentHash accumulates the hash of all processed segments.
	// Stored as string as in current tdf.go, could be []byte for efficiency
	aggregateSegmentHash string
	firstChunkSize       int64 // To store the size of the first chunk for DefaultSegmentSize
	firstChunkProcessed  bool
}

// NewStreamingTDFEncryptor initializes a new encryptor for creating a TDF in a streaming manner.
// The TDF structure (key access objects, policy) is prepared, and the payload stream is started.
// Chunks can then be added via AddChunk, and the TDF is finalized with Finalize.
func (s *SDK) NewStreamingTDFEncryptor(ctx context.Context, writer io.Writer, opts ...TDFOption) (*StreamingTDFEncryptor, error) {
	tdfConfig, err := newTDFConfig(opts...)
	if err != nil {
		return nil, fmt.Errorf("NewTDFConfig failed for streaming: %w", err)
	}

	// Autoconfigure KAS information if enabled (similar to CreateTDFContext)
	if tdfConfig.autoconfigure {
		var g granter
		g, err = s.newGranter(ctx, tdfConfig, nil) // Pass nil for existing error
		if err != nil {
			return nil, fmt.Errorf("newGranter failed for streaming: %w", err)
		}

		switch g.typ {
		case mappedFound:
			tdfConfig.kaoTemplate, err = g.resolveTemplate(uuidSplitIDGenerator)
		case grantsFound:
			tdfConfig.kaoTemplate = nil
			tdfConfig.splitPlan, err = g.plan(make([]string, 0), uuidSplitIDGenerator)
		case noKeysFound:
			baseKey, baseKeyErr := getBaseKeyFromWellKnown(ctx, *s)
			if baseKeyErr == nil {
				err = populateKasInfoFromBaseKey(baseKey, tdfConfig)
			} else {
				slog.DebugContext(ctx, "error getting base key for streaming, falling back to default kas", slog.Any("error", baseKeyErr))
				dk := s.defaultKases(tdfConfig)
				tdfConfig.kaoTemplate = nil
				tdfConfig.splitPlan, err = g.plan(dk, uuidSplitIDGenerator)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("failed to generate plan for streaming: %w", err)
		}
	}

	encryptor := &StreamingTDFEncryptor{
		sdk:               s,
		writer:            writer,
		tdfWriter:         archive.NewTDFWriter(writer),
		tdfConfig:         tdfConfig,
		collectedSegments: make([]Segment, 0),
		firstChunkSize:    -1, // Initialize to an invalid value
	}

	// Prepare the manifest structure (KAOs, policy, derive payloadKey)
	// This is similar to s.prepareManifest but adapted for streaming.
	// We pass a TDFObject equivalent for prepareManifest to populate.
	// In streaming, the manifest isn't fully complete until Finalize.
	tempTDFObject := &TDFObject{} // Used to satisfy prepareManifest requirements
	err = s.prepareManifest(ctx, tempTDFObject, *tdfConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare manifest for streaming: %w", err)
	}

	// Copy results from tempTDFObject
	encryptor.payloadKey = tempTDFObject.payloadKey
	encryptor.aesGcm = tempTDFObject.aesGcm
	encryptor.manifest = tempTDFObject.manifest // This is a partially filled manifest

	// Start the payload stream in the TDF archive
	if err := encryptor.tdfWriter.StartPayloadStream(); err != nil {
		return nil, fmt.Errorf("failed to start payload stream: %w", err)
	}

	return encryptor, nil
}

// AddChunk encrypts the given chunkData and appends it to the TDF payload stream.
// Information about the chunk (hash, size) is stored for final manifest generation.
func (e *StreamingTDFEncryptor) AddChunk(chunkData []byte) error {
	if e.aesGcm == nil {
		return fmt.Errorf("encryptor not properly initialized: aesGcm is nil")
	}
	if e.tdfWriter == nil {
		return fmt.Errorf("encryptor not properly initialized: tdfWriter is nil")
	}

	cipherData, err := e.aesGcm.Encrypt(chunkData)
	if err != nil {
		return fmt.Errorf("failed to encrypt chunk: %w", err)
	}

	if err := e.tdfWriter.AppendPayload(cipherData); err != nil {
		return fmt.Errorf("failed to append encrypted chunk to TDF: %w", err)
	}

	segmentSig, err := calculateSignature(cipherData, e.payloadKey[:],
		e.tdfConfig.segmentIntegrityAlgorithm, e.tdfConfig.useHex)
	if err != nil {
		return fmt.Errorf("failed to calculate segment signature: %w", err)
	}

	e.aggregateSegmentHash += segmentSig
	segmentInfo := Segment{
		Hash:          string(ocrypto.Base64Encode([]byte(segmentSig))),
		Size:          int64(len(chunkData)),
		EncryptedSize: int64(len(cipherData)),
	}
	e.collectedSegments = append(e.collectedSegments, segmentInfo)

	if !e.firstChunkProcessed {
		e.firstChunkSize = int64(len(chunkData))
		e.firstChunkProcessed = true
	}

	return nil
}

// Finalize completes the TDF creation process.
// It closes the payload stream, generates the complete manifest with all segment information
// and root signatures, appends the manifest to the TDF, and finalizes the archive.
// It returns the generated Manifest.
func (e *StreamingTDFEncryptor) Finalize(ctx context.Context) (*Manifest, error) {
	if e.tdfWriter == nil {
		return nil, fmt.Errorf("encryptor not properly initialized: tdfWriter is nil")
	}

	// Close the payload stream first
	if err := e.tdfWriter.ClosePayloadStream(); err != nil {
		return nil, fmt.Errorf("failed to close payload stream: %w", err)
	}

	// Now, finalize the manifest with collected segment information and root signature
	finalManifest := e.manifest // Start with the partially prepared manifest

	// Calculate root signature
	rootSignature, err := calculateSignature([]byte(e.aggregateSegmentHash), e.payloadKey[:],
		e.tdfConfig.integrityAlgorithm, e.tdfConfig.useHex)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate root signature: %w", err)
	}
	sig := string(ocrypto.Base64Encode([]byte(rootSignature)))
	finalManifest.EncryptionInformation.IntegrityInformation.RootSignature.Signature = sig

	integrityAlgStr := gmacIntegrityAlgorithm
	if e.tdfConfig.integrityAlgorithm == HS256 {
		integrityAlgStr = hmacIntegrityAlgorithm
	}
	finalManifest.EncryptionInformation.IntegrityInformation.RootSignature.Algorithm = integrityAlgStr

	// Set collected segments
	finalManifest.EncryptionInformation.IntegrityInformation.Segments = e.collectedSegments

	// Determine DefaultSegmentSize for the manifest
	// Use configured defaultSegmentSize, as per plan.
	defaultSegSize := e.tdfConfig.defaultSegmentSize
	finalManifest.EncryptionInformation.IntegrityInformation.DefaultSegmentSize = defaultSegSize
	finalManifest.EncryptionInformation.IntegrityInformation.DefaultEncryptedSegSize = defaultSegSize + gcmIvSize + aesBlockSize

	segIntegrityAlgStr := gmacIntegrityAlgorithm
	if e.tdfConfig.segmentIntegrityAlgorithm == HS256 {
		segIntegrityAlgStr = hmacIntegrityAlgorithm
	}
	finalManifest.EncryptionInformation.IntegrityInformation.SegmentHashAlgorithm = segIntegrityAlgStr
	finalManifest.EncryptionInformation.Method.IsStreamable = true // This is inherently streamable

	// Populate payload section of the manifest
	mimeType := e.tdfConfig.mimeType
	if mimeType == "" {
		mimeType = defaultMimeType
	}
	finalManifest.Payload.MimeType = mimeType
	finalManifest.Payload.Protocol = tdfAsZip
	finalManifest.Payload.Type = tdfZipReference
	finalManifest.Payload.URL = archive.TDFPayloadFileName
	finalManifest.Payload.IsEncrypted = true

	// Process assertions (reusing logic from tdf.go's CreateTDFContext)
	var signedAssertions []Assertion
	if e.tdfConfig.addDefaultAssertion {
		systemMeta, err := GetSystemMetadataAssertionConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get system metadata assertion config: %w", err)
		}
		e.tdfConfig.assertions = append(e.tdfConfig.assertions, systemMeta)
	}

	for _, assertionCfg := range e.tdfConfig.assertions {
		tmpAssertion := Assertion{
			ID:             assertionCfg.ID,
			Type:           assertionCfg.Type,
			Scope:          assertionCfg.Scope,
			Statement:      assertionCfg.Statement,
			AppliesToState: assertionCfg.AppliesToState,
		}

		hashOfAssertionAsHex, err := tmpAssertion.GetHash()
		if err != nil {
			return nil, fmt.Errorf("failed to get hash of assertion %s: %w", tmpAssertion.ID, err)
		}

		hashOfAssertion := make([]byte, hex.DecodedLen(len(hashOfAssertionAsHex)))
		_, err = hex.Decode(hashOfAssertion, hashOfAssertionAsHex)
		if err != nil {
			return nil, fmt.Errorf("error decoding hex string for assertion %s: %w", tmpAssertion.ID, err)
		}

		var completeHashBuilder strings.Builder
		completeHashBuilder.WriteString(e.aggregateSegmentHash)
		if e.tdfConfig.useHex { // Check if old TDF version hex encoding should be used
			completeHashBuilder.Write(hashOfAssertionAsHex)
		} else {
			completeHashBuilder.Write(hashOfAssertion)
		}
		encodedPayloadAndAssertionHash := ocrypto.Base64Encode([]byte(completeHashBuilder.String()))

		assertionSigningKey := AssertionKey{
			Alg: AssertionKeyAlgHS256,
			Key: e.payloadKey[:],
		}
		if !assertionCfg.SigningKey.IsEmpty() {
			assertionSigningKey = assertionCfg.SigningKey
		}

		if err := tmpAssertion.Sign(string(hashOfAssertionAsHex), string(encodedPayloadAndAssertionHash), assertionSigningKey); err != nil {
			return nil, fmt.Errorf("failed to sign assertion %s: %w", tmpAssertion.ID, err)
		}
		signedAssertions = append(signedAssertions, tmpAssertion)
	}
	finalManifest.Assertions = signedAssertions

	// Marshal the finalized manifest to JSON
	manifestAsStr, err := json.Marshal(finalManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal final manifest: %w", err)
	}

	// Append the manifest to the TDF archive
	if err := e.tdfWriter.AppendManifest(string(manifestAsStr)); err != nil {
		return nil, fmt.Errorf("failed to append manifest to TDF: %w", err)
	}

	// Finish writing the TDF archive (writes central directory, etc.)
	_, err = e.tdfWriter.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish TDF archive: %w", err)
	}

	e.manifest = finalManifest // Store the fully finalized manifest
	return &e.manifest, nil
}

// Manifest returns the final manifest after successful finalization.
// Returns an empty manifest if Finalize() has not been called or was unsuccessful.
func (e *StreamingTDFEncryptor) Manifest() Manifest {
	return e.manifest
}
