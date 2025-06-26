# Go SDK: Indeterministic Streaming TDF Encryption

The OpenTDF Go SDK provides a way to encrypt data into a TDF (TDF) format when the total size of the input data is not known in advance, or when it's beneficial to process data in chunks without holding the entire dataset in memory. This is referred to as indeterministic streaming.

## Use Cases

This feature is particularly useful for:

*   **Encrypting large files:** When dealing with files that are too large to fit comfortably in memory, you can read and encrypt them chunk by chunk.
*   **Receiving data streams:** If you are receiving data from a network stream or another source where the end of the data is not known upfront.
*   **Reducing memory footprint:** Processing data in smaller, manageable chunks can significantly reduce the memory requirements of your application during encryption.

## How It Works

The streaming encryption process involves the following steps:

1.  **Initialization:** A `StreamingTDFEncryptor` is created. During this phase, the necessary policy information is processed, Key Access Objects (KAOs) are prepared, and the encryption keys for the TDF payload are derived. The TDF archive (which is a ZIP file) is set up to expect payload data.
2.  **Chunk Processing:** Arbitrary-sized chunks of plaintext data are provided to the encryptor. Each chunk is encrypted immediately and appended to the TDF's payload stream. Information about each segment (its original size, encrypted size, and hash) is collected by the encryptor.
3.  **Finalization:** Once all data chunks have been processed, the TDF is finalized. This involves:
    *   Calculating the root integrity signature for the entire payload from the hashes of individual segments.
    *   Assembling the complete TDF manifest, including all segment details, key access information, policy, and the root signature.
    *   Writing the manifest to the TDF archive.
    *   Finalizing the TDF (ZIP) archive structure.

The key aspect is that the manifest, which contains metadata about all the encrypted segments, is only written *after* all payload data has been processed.

## Usage Example

Here's how to use the `StreamingTDFEncryptor`:

```go
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/opentdf/platform/sdk"
	"github.com/opentdf/platform/lib/ocrypto" // For ocrypto.RSA2048Key etc.
)

func main() {
	// Replace with your actual platform endpoint and authentication options
	platformEndpoint := "http://localhost:8080"
	kasURL := platformEndpoint + "/kas" // Example KAS URL

	// Initialize the SDK client
	// Ensure you have appropriate opts for authentication if required by your platform setup
	opts := []sdk.Option{
		// Example: sdk.WithClientCredentials("your-client-id", "your-client-secret", nil),
	}
	sdkClient, err := sdk.New(platformEndpoint, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize SDK: %v\n", err)
		return
	}

	// Prepare an output buffer or file writer for the TDF
	var tdfOutput bytes.Buffer // Or use os.Create("sensitive.tdf") for a file

	// Data attribute(s) for the TDF
	dataAttributes := []string{"https://example.com/attr/mystream/value/secret"}

	// KAS information
	kasInfos := []sdk.KASInfo{
		{URL: kasURL, PublicKey: "" /* SDK will fetch if empty */, KID: ""},
	}

	// Initialize the StreamingTDFEncryptor
	// Using context.Background() for simplicity in this example
	encryptor, err := sdkClient.NewStreamingTDFEncryptor(context.Background(), &tdfOutput,
		sdk.WithDataAttributes(dataAttributes...),
		sdk.WithKasInformation(kasInfos...),
		sdk.WithWrappingKeyAlg(ocrypto.RSA2048Key), // Specify a key wrapping algorithm
		// sdk.WithMetaData("Optional metadata string"),
		// sdk.WithMimeType("application/octet-stream"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create StreamingTDFEncryptor: %v\n", err)
		return
	}

	// Define your data chunks
	chunks := [][]byte{
		[]byte("This is the first part of the secret message. "),
		[]byte("Here comes the second part, adding more data. "),
		[]byte("And finally, the concluding segment of this stream."),
	}

	// Add each chunk to the encryptor
	for i, chunk := range chunks {
		if err := encryptor.AddChunk(chunk); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add chunk %d: %v\n", i+1, err)
			return
		}
		fmt.Printf("Added chunk %d\n", i+1)
	}

	// Finalize the TDF creation
	finalManifest, err := encryptor.Finalize(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to finalize TDF: %v\n", err)
		return
	}

	fmt.Printf("TDF encryption successful. Manifest UUID: %s\n", finalManifest.EncryptionInformation.Policy) // Policy UUID
	fmt.Printf("TDF output size: %d bytes\n", tdfOutput.Len())

	// At this point, tdfOutput contains the complete TDF data.
	// You can write it to a file, upload it, etc.
	// For example, to write to a file:
	// err = os.WriteFile("sensitive_streamed.tdf", tdfOutput.Bytes(), 0644)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Failed to write TDF to file: %v\n", err)
	// }

	// To verify, you can decrypt it (example assumes you have a decrypt function or use the SDK's decrypt capabilities)
	// For decryption, you would use sdkClient.LoadTDF(&tdfOutput) and then read from it.
}
```

## Manifest Details

When using streaming encryption:

*   The `DefaultSegmentSize` and `DefaultEncryptedSegSize` fields in the TDF manifest are populated based on the `defaultSegmentSize` value provided in the `TDFConfig` (via `sdk.WithSegmentSize` option, or the SDK's internal default if not specified). This value might not reflect the actual size of each chunk if they vary, but it provides a consistent value for the manifest structure.
*   The `Segments` array within the manifest will accurately list each processed chunk with its actual original size, encrypted size, and hash.
*   The `Method.IsStreamable` field in the manifest will be set to `true`.

This approach ensures that even with indeterministic input, the resulting TDF is valid and compatible with standard TDF readers.
```
