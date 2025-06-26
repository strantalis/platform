package archive

import (
	"fmt"
	"io"
)

type TDFWriter struct {
	archiveWriter *Writer
}

// NewTDFWriter Create tdf writer instance.
func NewTDFWriter(writer io.Writer) *TDFWriter {
	tdfWriter := TDFWriter{}
	tdfWriter.archiveWriter = NewWriter(writer)

	return &tdfWriter
}

// StartPayloadStream prepares the TDF archive for writing the payload when its total size is not yet known.
// It configures the underlying archive writer to use Zip64 for streaming if necessary and
// adds the header for the payload file with an unknown size.
func (tdfWriter *TDFWriter) StartPayloadStream() error {
	// For streaming, it's safer to enable Zip64 as the final size is unknown
	// and could exceed Zip32 limits. The underlying archive.Writer will still
	// only use Zip64 features if actually needed for a given entry or the archive as a whole.
	tdfWriter.archiveWriter.EnableZip64()
	return tdfWriter.archiveWriter.AddHeader(TDFPayloadFileName, SizeUnknown)
}

// ClosePayloadStream finalizes the payload entry in the TDF archive.
// This must be called after all payload data has been written via AppendPayload.
func (tdfWriter *TDFWriter) ClosePayloadStream() error {
	return tdfWriter.archiveWriter.CloseFileEntry()
}

// AppendManifest adds the manifest to the TDF archive.
// This must be called *after* ClosePayloadStream has been called and all payload data written.
func (tdfWriter *TDFWriter) AppendManifest(manifest string) error {
	// Add header for the manifest file (size is known)
	err := tdfWriter.archiveWriter.AddHeader(TDFManifestFileName, int64(len(manifest)))
	if err != nil {
		return fmt.Errorf("failed to add manifest header: %w", err)
	}

	// Add manifest data
	err = tdfWriter.archiveWriter.AddData([]byte(manifest))
	if err != nil {
		return fmt.Errorf("failed to add manifest data: %w", err)
	}

	// Close the manifest file entry
	return tdfWriter.archiveWriter.CloseFileEntry()
}

// AppendPayload adds a chunk of payload data to the TDF archive.
// StartPayloadStream must have been called before this.
func (tdfWriter *TDFWriter) AppendPayload(data []byte) error {
	return tdfWriter.archiveWriter.AddData(data)
}

// Finish Finished adding all the files in zip archive.
func (tdfWriter *TDFWriter) Finish() (int64, error) {
	return tdfWriter.archiveWriter.Finish()
}
