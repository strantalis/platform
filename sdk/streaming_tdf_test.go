package sdk

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	mathRand "math/rand"
	"testing"

	"github.com/opentdf/platform/lib/ocrypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamingTDF_EncryptDecrypt_RandomChunks(t *testing.T) {
	sdk, err := New(platformEndpoint, opts...)
	require.NoError(t, err)

	// 1. Prepare input data
	originalDataSize := 5 * 1024 * 1024 // 5MB
	originalData := make([]byte, originalDataSize)
	_, err = rand.Read(originalData)
	require.NoError(t, err)

	var tdfBuffer bytes.Buffer

	// 2. Encrypt using StreamingTDFEncryptor
	encryptor, err := sdk.NewStreamingTDFEncryptor(context.Background(), &tdfBuffer,
		WithDataAttributes("https://example.com/attr/InStream/value/Test"),
		WithKasInformation(KASInfo{URL: kasURL}), // Assuming kasURL is a globally defined test KAS
		WithWrappingKeyAlg(ocrypto.RSA2048Key), // Specify a wrapping key algorithm
	)
	require.NoError(t, err)
	require.NotNil(t, encryptor)

	// Add data in random-sized chunks
	bytesSent := 0
	minChunkSize := 1024        // 1KB
	maxChunkSize := 512 * 1024 // 512KB

	for bytesSent < originalDataSize {
		chunkSize := mathRand.Intn(maxChunkSize-minChunkSize+1) + minChunkSize
		if bytesSent+chunkSize > originalDataSize {
			chunkSize = originalDataSize - bytesSent
		}
		chunk := originalData[bytesSent : bytesSent+chunkSize]
		err = encryptor.AddChunk(chunk)
		require.NoError(t, err, "Failed to add chunk at offset %d with size %d", bytesSent, chunkSize)
		bytesSent += chunkSize
	}

	finalManifest, err := encryptor.Finalize(context.Background())
	require.NoError(t, err)
	require.NotNil(t, finalManifest)

	// 3. Manifest Verification (Basic Checks)
	assert.NotEmpty(t, finalManifest.EncryptionInformation.IntegrityInformation.Segments)
	assert.Equal(t, defaultSegmentSize, finalManifest.EncryptionInformation.IntegrityInformation.DefaultSegmentSize) // Check if it used the config default
	assert.NotEmpty(t, finalManifest.EncryptionInformation.IntegrityInformation.RootSignature.Signature)
	assert.Equal(t, archive.TDFPayloadFileName, finalManifest.Payload.URL)

	var totalPlaintextInManifest int64
	for _, seg := range finalManifest.EncryptionInformation.IntegrityInformation.Segments {
		totalPlaintextInManifest += seg.Size
	}
	assert.Equal(t, int64(originalDataSize), totalPlaintextInManifest, "Total segment size in manifest should match original data size")

	// 4. Decrypt and Verify
	tdfReader, err := sdk.LoadTDF(bytes.NewReader(tdfBuffer.Bytes()))
	require.NoError(t, err)

	// Perform initial unwrap
	err = tdfReader.Init(context.Background())
	require.NoError(t, err, "tdfReader.Init (unwrap) failed")

	decryptedData, err := io.ReadAll(tdfReader)
	require.NoError(t, err)

	assert.Equal(t, originalDataSize, len(decryptedData), "Decrypted data size mismatch")
	assert.Equal(t, originalData, decryptedData, "Decrypted data content mismatch")

	// Test ReadAt for good measure
	readAtBuffer := make([]byte, 100)
	offset := int64(originalDataSize / 2)
	n, err := tdfReader.ReadAt(readAtBuffer, offset)
	require.NoError(t, err, "ReadAt failed")
	assert.Equal(t, 100, n)
	assert.Equal(t, originalData[offset:offset+100], readAtBuffer[:n])

}

// Minimal KAS setup for tests - replace with actual test KAS URL if available from existing test setup
// For now, using a placeholder. This test will likely need a running KAS.
// If existing tests in sdk_test.go setup a mock KAS, that would be ideal to reuse.
var kasURL = "http://localhost:8080/kas" // Placeholder - this needs to be a functional KAS for the test to pass E2E

// TODO: If there's a common test setup for SDK (like in sdk_test.go),
// try to reuse KASInfo and platformEndpoint from there.
// For now, this is a self-contained test.
// The `opts` variable used in NewSDK would also come from a shared test setup or be defined here.
// var opts = []Option{WithClientCredentials("test-client", "test-secret", nil)} // Example, adjust as needed

func TestStreamingTDF_EmptyPayload(t *testing.T) {
	sdk, err := New(platformEndpoint, opts...)
	require.NoError(t, err)

	var tdfBuffer bytes.Buffer
	encryptor, err := sdk.NewStreamingTDFEncryptor(context.Background(), &tdfBuffer,
		WithDataAttributes("https://example.com/attr/EmptyStream/value/Test"),
		WithKasInformation(KASInfo{URL: kasURL}),
		WithWrappingKeyAlg(ocrypto.RSA2048Key),
	)
	require.NoError(t, err)

	// No AddChunk calls

	finalManifest, err := encryptor.Finalize(context.Background())
	require.NoError(t, err)
	require.NotNil(t, finalManifest)

	assert.Empty(t, finalManifest.EncryptionInformation.IntegrityInformation.Segments, "Segments list should be empty for empty payload")
	// assert.Equal(t, int64(0), finalManifest.EncryptionInformation.IntegrityInformation.DefaultSegmentSize) // Or defaultSegmentSize from config
	assert.Equal(t, defaultSegmentSize, finalManifest.EncryptionInformation.IntegrityInformation.DefaultSegmentSize)
	assert.NotEmpty(t, finalManifest.EncryptionInformation.IntegrityInformation.RootSignature.Signature) // Root sig is over empty aggregate hash

	// Decrypt and Verify
	tdfReader, err := sdk.LoadTDF(bytes.NewReader(tdfBuffer.Bytes()))
	require.NoError(t, err)
	err = tdfReader.Init(context.Background())
	require.NoError(t, err, "tdfReader.Init (unwrap) failed for empty payload")


	decryptedData, err := io.ReadAll(tdfReader)
	require.NoError(t, err)
	assert.Empty(t, decryptedData, "Decrypted data should be empty")
}

func TestStreamingTDF_OneChunk(t *testing.T) {
	sdk, err := New(platformEndpoint, opts...)
	require.NoError(t, err)

	originalData := []byte("this is a single chunk of data")
	var tdfBuffer bytes.Buffer

	encryptor, err := sdk.NewStreamingTDFEncryptor(context.Background(), &tdfBuffer,
		WithDataAttributes("https://example.com/attr/OneChunk/value/Test"),
		WithKasInformation(KASInfo{URL: kasURL}),
		WithWrappingKeyAlg(ocrypto.RSA2048Key),
	)
	require.NoError(t, err)

	err = encryptor.AddChunk(originalData)
	require.NoError(t, err)

	finalManifest, err := encryptor.Finalize(context.Background())
	require.NoError(t, err)
	require.NotNil(t, finalManifest)

	require.Len(t, finalManifest.EncryptionInformation.IntegrityInformation.Segments, 1)
	assert.Equal(t, int64(len(originalData)), finalManifest.EncryptionInformation.IntegrityInformation.Segments[0].Size)

	// Decrypt and Verify
	tdfReader, err := sdk.LoadTDF(bytes.NewReader(tdfBuffer.Bytes()))
	require.NoError(t, err)
	err = tdfReader.Init(context.Background())
	require.NoError(t, err, "tdfReader.Init (unwrap) failed for single chunk")

	decryptedData, err := io.ReadAll(tdfReader)
	require.NoError(t, err)
	assert.Equal(t, originalData, decryptedData)
}

// Note: These tests currently use a placeholder `kasURL` and `opts`.
// For them to run successfully in a CI environment, they would need to be integrated
// with the existing test infrastructure, which might involve:
// - A running KAS instance (mock or real).
// - Proper configuration for the SDK client (authentication, platform endpoint).
// The `platformEndpoint` variable also needs to be available, likely from a shared test setup.
// If sdk_test.go has a TestMain or setup functions, those should be leveraged.

// Assuming platformEndpoint and opts are defined globally for tests, like in sdk_test.go
// If not, they need to be defined or passed appropriately.
// var platformEndpoint = "http://localhost:8080" // Example
// var opts = []Option{WithClientCredentials("client", "secret", nil)} // Example
