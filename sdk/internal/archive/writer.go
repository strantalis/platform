//nolint:mnd // pkzip magics and lengths are inlined for clarity
package archive

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"time"
)

const (
	// SizeUnknown is a special value for AddHeader's size argument
	// to indicate that the file size is not known at the time of header creation.
	// The Data Descriptor will be used to record the size later.
	SizeUnknown = -1
)

// https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT
// https://rzymek.github.io/post/excel-zip64/
// Overall .ZIP file format:
//   [local file header 1]
//   [file data 1]
//   [ext 1]
//   [data descriptor 1]
//   .
//   .
//   .
//   [local file header n]
//   [file data n]
//   [ext n]
//   [data descriptor n]
//   [central directory header 1]
//   .
//   .
//   .
//   [central directory header n]
//   [zip64 end of central directory record]
//   [zip64 end of central directory locator]
//   [end of central directory record]

// Usage of IArchiveWriter interface:
//
// NOTE: Make sure write the largest file first so the implementation can decide zip32 vs zip64

type WriteState int

const (
	Initial WriteState = iota
	Appending
	// Finished // This state is effectively replaced by calling CloseFileEntry
)

type FileInfo struct {
	filename          string
	headerOffset      uint64 // Offset of the local file header
	accumulatedSize   int64  // Actual size of the file data written so far
	declaredSize      int64  // Size declared in AddHeader, or SizeUnknown
	currentCRC32      uint32
	fileTime          uint16
	fileDate          uint16
	flag              uint16
	isSizeKnown       bool
	useZip64ExtraField bool // True if this specific file needs a Zip64 extended info in its local header
}

type Writer struct {
	writer                io.Writer
	currentOffset         uint64 // Tracks global offset in the archive
	lastOffsetCDFileHeader uint64
	activeFileInfo        *FileInfo // Information about the file currently being written
	fileInfoEntries       []FileInfo // Stores finalized FileInfo for central directory
	writeState            WriteState
	forceZip64            bool // True if zip64 is forced for the entire archive
	totalBytes            int64
}

// NewWriter Create tdf3 writer instance.
func NewWriter(writer io.Writer) *Writer {
	archiveWriter := Writer{}

	archiveWriter.writer = writer
	archiveWriter.writeState = Initial
	archiveWriter.currentOffset = 0
	archiveWriter.lastOffsetCDFileHeader = 0
	archiveWriter.fileInfoEntries = make([]FileInfo, 0)

	return &archiveWriter
}

// EnableZip64 Enable zip 64 for the entire archive.
func (writer *Writer) EnableZip64() {
	writer.forceZip64 = true
}

// AddHeader prepares to write a new file to the archive.
// size is the size of the file. If the size is not known beforehand (e.g., for streaming),
// pass SizeUnknown. In this case, a data descriptor will be written after the file data.
func (writer *Writer) AddHeader(filename string, size int64) error {
	if writer.activeFileInfo != nil {
		return fmt.Errorf("writer: previous file entry for '%s' must be closed with CloseFileEntry before starting a new one", writer.activeFileInfo.filename)
	}

	writer.activeFileInfo = &FileInfo{
		filename:     filename,
		declaredSize: size,
		headerOffset: writer.currentOffset,
		isSizeKnown:  size != SizeUnknown,
		currentCRC32: crc32.Checksum([]byte(""), crc32.MakeTable(crc32.IEEE)),
	}
	writer.activeFileInfo.fileTime, writer.activeFileInfo.fileDate = writer.getTimeDateUnMSDosFormat()
	writer.activeFileInfo.flag = 0x08 // Always use data descriptor for consistency

	localFileHeader := LocalFileHeader{
		Signature:           fileHeaderSignature,
		Version:             zipVersion,
		GeneralPurposeBitFlag: writer.activeFileInfo.flag,
		CompressionMethod:   0, // no compression
		LastModifiedTime:    writer.activeFileInfo.fileTime,
		LastModifiedDate:    writer.activeFileInfo.fileDate,
		Crc32:               0, // Will be in data descriptor
		FilenameLength:      uint16(len(writer.activeFileInfo.filename)),
	}

	fileNeedsZip64 := writer.forceZip64 || (writer.activeFileInfo.isSizeKnown && size >= zip64MagicVal) || (!writer.activeFileInfo.isSizeKnown && writer.forceZip64)
	writer.activeFileInfo.useZip64ExtraField = fileNeedsZip64


	if fileNeedsZip64 {
		localFileHeader.CompressedSize = zip64MagicVal   // Indicate Zip64 in local header
		localFileHeader.UncompressedSize = zip64MagicVal // Indicate Zip64 in local header
		localFileHeader.ExtraFieldLength = zip64ExtendedLocalInfoExtraFieldSize
	} else {
		// Sizes are 0 because they will be in the data descriptor
		localFileHeader.CompressedSize = 0
		localFileHeader.UncompressedSize = 0
		localFileHeader.ExtraFieldLength = 0
	}

	// Write localFileHeader
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, localFileHeader); err != nil {
		return fmt.Errorf("binary.Write localFileHeader failed: %w", err)
	}
	if err := writer.writeData(buf.Bytes()); err != nil {
		return fmt.Errorf("writer.writeData for localFileHeader failed: %w", err)
	}
	writer.currentOffset += uint64(localFileHeaderSize)

	// Write the file name
	if err := writer.writeData([]byte(writer.activeFileInfo.filename)); err != nil {
		return fmt.Errorf("writer.writeData for filename failed: %w", err)
	}
	writer.currentOffset += uint64(len(writer.activeFileInfo.filename))

	if fileNeedsZip64 {
		zip64Extra := Zip64ExtendedLocalInfoExtraField{
			Signature: zip64ExternalID,
			Size:      zip64ExtendedLocalInfoExtraFieldSize - 4, // Size of this extra field block minus sig and size fields
		}
		if writer.activeFileInfo.isSizeKnown {
			zip64Extra.OriginalSize = uint64(size)
			zip64Extra.CompressedSize = uint64(size) // Assuming no compression
		} else {
			// Sizes are unknown, will be in data descriptor. ZIP spec says these should be present.
			// Some interpretations suggest they could be zero if unknown, but safest to include them as per spec example for streaming.
			// However, since we are using data descriptor, these values in local zip64 extra field are often ignored by readers.
			// Let's use 0 to signify unknown here, as the central directory will hold the true values.
			zip64Extra.OriginalSize = 0
			zip64Extra.CompressedSize = 0
		}

		buf.Reset()
		if err := binary.Write(buf, binary.LittleEndian, zip64Extra); err != nil {
			return fmt.Errorf("binary.Write zip64ExtendedLocalInfoExtraField failed: %w", err)
		}
		if err := writer.writeData(buf.Bytes()); err != nil {
			return fmt.Errorf("writer.writeData for zip64ExtendedLocalInfoExtraField failed: %w", err)
		}
		writer.currentOffset += uint64(zip64ExtendedLocalInfoExtraFieldSize)
	}
	writer.writeState = Appending
	return nil
}

// AddData adds data to the current file in the zip archive.
// AddHeader must be called before AddData.
func (writer *Writer) AddData(data []byte) error {
	if writer.activeFileInfo == nil {
		return fmt.Errorf("writer: AddHeader must be called before AddData")
	}
	if writer.writeState != Appending {
		return fmt.Errorf("writer: file entry is not in appending state; current file: %s", writer.activeFileInfo.filename)
	}

	if err := writer.writeData(data); err != nil {
		return fmt.Errorf("io.Writer.Write failed: %w", err)
	}

	writer.activeFileInfo.currentCRC32 = crc32.Update(writer.activeFileInfo.currentCRC32, crc32.MakeTable(crc32.IEEE), data)
	writer.activeFileInfo.accumulatedSize += int64(len(data))
	writer.currentOffset += uint64(len(data))

	return nil
}

// CloseFileEntry finalizes the current file entry by writing its data descriptor.
// This must be called after all data for the current file has been written with AddData.
func (writer *Writer) CloseFileEntry() error {
	if writer.activeFileInfo == nil {
		return fmt.Errorf("writer: no active file entry to close")
	}
	if writer.writeState != Appending {
		return fmt.Errorf("writer: file entry is not in appending state to close; current file: %s", writer.activeFileInfo.filename)
	}

	// Finalize size if it was unknown
	if !writer.activeFileInfo.isSizeKnown {
		writer.activeFileInfo.declaredSize = writer.activeFileInfo.accumulatedSize
	} else if writer.activeFileInfo.accumulatedSize != writer.activeFileInfo.declaredSize {
		return fmt.Errorf("writer: accumulated size %d does not match declared size %d for file %s",
			writer.activeFileInfo.accumulatedSize, writer.activeFileInfo.declaredSize, writer.activeFileInfo.filename)
	}

	actualSize := writer.activeFileInfo.accumulatedSize
	fileNeedsZip64 := writer.forceZip64 || actualSize >= zip64MagicVal

	if fileNeedsZip64 {
		dataDescriptor := Zip64DataDescriptor{
			Signature:        dataDescriptorSignature,
			Crc32:            writer.activeFileInfo.currentCRC32,
			CompressedSize:   uint64(actualSize),
			UncompressedSize: uint64(actualSize),
		}
		buf := new(bytes.Buffer)
		if err := binary.Write(buf, binary.LittleEndian, dataDescriptor); err != nil {
			return fmt.Errorf("binary.Write Zip64DataDescriptor failed: %w", err)
		}
		if err := writer.writeData(buf.Bytes()); err != nil {
			return fmt.Errorf("writer.writeData for Zip64DataDescriptor failed: %w", err)
		}
		writer.currentOffset += uint64(zip64DataDescriptorSize)
	} else {
		dataDescriptor := Zip32DataDescriptor{
			Signature:        dataDescriptorSignature,
			Crc32:            writer.activeFileInfo.currentCRC32,
			CompressedSize:   uint32(actualSize),
			UncompressedSize: uint32(actualSize),
		}
		buf := new(bytes.Buffer)
		if err := binary.Write(buf, binary.LittleEndian, dataDescriptor); err != nil {
			return fmt.Errorf("binary.Write Zip32DataDescriptor failed: %w", err)
		}
		if err := writer.writeData(buf.Bytes()); err != nil {
			return fmt.Errorf("writer.writeData for Zip32DataDescriptor failed: %w", err)
		}
		writer.currentOffset += uint64(zip32DataDescriptorSize)
	}

	// Store finalized FileInfo for central directory
	// Note: The 'offset' in FileInfo for central directory is the local header offset.
	// The 'size' is the actual accumulated size.
	finalizedInfo := FileInfo{
		filename:     writer.activeFileInfo.filename,
		headerOffset: writer.activeFileInfo.headerOffset,
		accumulatedSize:  actualSize, // Store the actual size
		declaredSize: writer.activeFileInfo.declaredSize, // Keep declared for reference, though accumulated is authoritative now
		currentCRC32: writer.activeFileInfo.currentCRC32,
		fileTime:     writer.activeFileInfo.fileTime,
		fileDate:     writer.activeFileInfo.fileDate,
		flag:         writer.activeFileInfo.flag,
		isSizeKnown:  true, // Size is now known
		useZip64ExtraField: writer.activeFileInfo.useZip64ExtraField || fileNeedsZip64, // It might need zip64 in CD even if local didn't, or vice-versa if size was unknown
	}
	writer.fileInfoEntries = append(writer.fileInfoEntries, finalizedInfo)
	writer.activeFileInfo = nil // Reset active file info
	writer.writeState = Initial // Ready for new header or finish

	return nil
}


// Finish Finished adding all the files in zip archive.
func (writer *Writer) Finish() (int64, error) {
	err := writer.writeCentralDirectory()
	if err != nil {
		return writer.totalBytes, err
	}

	err = writer.writeEndOfCentralDirectory()
	if err != nil {
		return writer.totalBytes, fmt.Errorf("io.Writer.Write failed: %w", err)
	}

	return writer.totalBytes, nil
}

// WriteZip64EndOfCentralDirectory write the zip64 end of central directory record struct to the archive.
func (writer *Writer) WriteZip64EndOfCentralDirectory() error {
	zip64EndOfCDRecord := Zip64EndOfCDRecord{}
	zip64EndOfCDRecord.Signature = zip64EndOfCDSignature
	zip64EndOfCDRecord.RecordSize = zip64EndOfCDRecordSize - 12
	zip64EndOfCDRecord.VersionMadeBy = zipVersion
	zip64EndOfCDRecord.VersionToExtract = zipVersion
	zip64EndOfCDRecord.DiskNumber = 0
	zip64EndOfCDRecord.StartDiskNumber = 0
	zip64EndOfCDRecord.NumberOfCDRecordEntries = uint64(len(writer.fileInfoEntries))
	zip64EndOfCDRecord.TotalCDRecordEntries = uint64(len(writer.fileInfoEntries))
	zip64EndOfCDRecord.CentralDirectorySize = writer.lastOffsetCDFileHeader - writer.currentOffset
	zip64EndOfCDRecord.StartingDiskCentralDirectoryOffset = writer.currentOffset

	// write the zip64 end of central directory record struct
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, zip64EndOfCDRecord)
	if err != nil {
		return fmt.Errorf("binary.Write failed: %w", err)
	}

	err = writer.writeData(buf.Bytes())
	if err != nil {
		return fmt.Errorf("io.Writer.Write failed: %w", err)
	}

	return nil
}

// WriteZip64EndOfCentralDirectoryLocator write the zip64 end of central directory locator struct
// to the archive.
func (writer *Writer) WriteZip64EndOfCentralDirectoryLocator() error {
	zip64EndOfCDRecordLocator := Zip64EndOfCDRecordLocator{}
	zip64EndOfCDRecordLocator.Signature = zip64EndOfCDLocatorSignature
	zip64EndOfCDRecordLocator.CDStartDiskNumber = 0
	zip64EndOfCDRecordLocator.CDOffset = writer.lastOffsetCDFileHeader
	zip64EndOfCDRecordLocator.NumberOfDisks = 1

	// write the zip64 end of central directory locator struct
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, zip64EndOfCDRecordLocator)
	if err != nil {
		return fmt.Errorf("binary.Write failed: %w", err)
	}

	err = writer.writeData(buf.Bytes())
	if err != nil {
		return fmt.Errorf("io.Writer.Write failed: %w", err)
	}

	return nil
}

// GetTimeDateUnMSDosFormat Get the time and date in MSDOS format.
const defaultSecondValue = 29

const monthShift = 5

const baseYear = 80

const halfSecond = 2

func (writer *Writer) getTimeDateUnMSDosFormat() (uint16, uint16) {
	t := time.Now().UTC()
	timeInDos := t.Hour()<<11 | t.Minute()<<5 | int(math.Max(float64(t.Second()/halfSecond), float64(defaultSecondValue)))
	dateInDos := (t.Year()-baseYear)<<9 | int((t.Month()+1)<<monthShift) | t.Day()
	return uint16(timeInDos), uint16(dateInDos)
}

func (writer *Writer) writeData(data []byte) error {
	n, err := writer.writer.Write(data)
	if err != nil {
		return err
	}

	writer.totalBytes += int64(n)
	return nil
}

// WriteCentralDirectory write central directory struct into archive.
func (writer *Writer) writeCentralDirectory() error {
	cdStartOffset := writer.currentOffset // Offset where the first CD header will be written
	writer.lastOffsetCDFileHeader = writer.currentOffset // Keep track of end of CD for Zip64 EOCD record

	for _, fileInfo := range writer.fileInfoEntries {
		cdFileHeader := CDFileHeader{
			Signature:              centralDirectoryHeaderSignature,
			VersionCreated:         zipVersion,
			VersionNeeded:          zipVersion,
			GeneralPurposeBitFlag:  fileInfo.flag,
			CompressionMethod:      0, // No compression
			LastModifiedTime:       fileInfo.fileTime,
			LastModifiedDate:       fileInfo.fileDate,
			Crc32:                  fileInfo.currentCRC32,
			FilenameLength:         uint16(len(fileInfo.filename)),
			FileCommentLength:      0,
			DiskNumberStart:        0,
			InternalFileAttributes: 0,
			ExternalFileAttributes: 0,
		}

		isZip64Entry := writer.forceZip64 || fileInfo.accumulatedSize >= zip64MagicVal || fileInfo.headerOffset >= zip64MagicVal

		if isZip64Entry {
			cdFileHeader.CompressedSize = zip64MagicVal
			cdFileHeader.UncompressedSize = zip64MagicVal
			cdFileHeader.LocalHeaderOffset = zip64MagicVal
			cdFileHeader.ExtraFieldLength = zip64ExtendedInfoExtraFieldSize
		} else {
			cdFileHeader.CompressedSize = uint32(fileInfo.accumulatedSize)
			cdFileHeader.UncompressedSize = uint32(fileInfo.accumulatedSize)
			cdFileHeader.LocalHeaderOffset = uint32(fileInfo.headerOffset)
			cdFileHeader.ExtraFieldLength = 0
		}


		// Write central directory file header struct
		buf := new(bytes.Buffer)
		if err := binary.Write(buf, binary.LittleEndian, cdFileHeader); err != nil {
			return fmt.Errorf("binary.Write CDFileHeader for %s failed: %w", fileInfo.filename, err)
		}
		if err := writer.writeData(buf.Bytes()); err != nil {
			return fmt.Errorf("writer.writeData for CDFileHeader of %s failed: %w", fileInfo.filename, err)
		}
		writer.currentOffset += uint64(cdFileHeaderSize)


		// Write the filename
		if err := writer.writeData([]byte(fileInfo.filename)); err != nil {
			return fmt.Errorf("writer.writeData for filename in CD of %s failed: %w", fileInfo.filename, err)
		}
		writer.currentOffset += uint64(len(fileInfo.filename))

		if isZip64Entry {
			zip64Extra := Zip64ExtendedInfoExtraField{
				Signature:             zip64ExternalID,
				Size:                  zip64ExtendedInfoExtraFieldSize - 4, // Size of this specific extra field part
				OriginalSize:          uint64(fileInfo.accumulatedSize),
				CompressedSize:        uint64(fileInfo.accumulatedSize), // Assuming no compression
				LocalFileHeaderOffset: fileInfo.headerOffset,
				// DiskStartNumber is not included as per spec for this field (it's 0 for this implementation)
			}

			buf.Reset()
			if err := binary.Write(buf, binary.LittleEndian, zip64Extra); err != nil {
				return fmt.Errorf("binary.Write Zip64ExtendedInfoExtraField for %s failed: %w", fileInfo.filename, err)
			}
			if err := writer.writeData(buf.Bytes()); err != nil {
				return fmt.Errorf("writer.writeData for Zip64ExtendedInfoExtraField of %s failed: %w", fileInfo.filename, err)
			}
			writer.currentOffset += uint64(zip64ExtendedInfoExtraFieldSize)
		}
	}
	// After loop, writer.currentOffset is the offset of the end of the last CD entry (which is start of EOCD)
	writer.lastOffsetCDFileHeader = writer.currentOffset // This is the start of EOCD / end of CD
	writer.currentOffset = cdStartOffset // Reset currentOffset to the start of CD for EOCD calculations
	return nil
}

// writeEndOfCentralDirectory write end of central directory struct into archive.
func (writer *Writer) writeEndOfCentralDirectory() error {
	// currentOffset is currently start of CD, lastOffsetCDFileHeader is end of CD / start of EOCD
	cdSize := writer.lastOffsetCDFileHeader - writer.currentOffset
	cdStartActualOffset := writer.currentOffset // This is the actual starting offset of CD

	archiveIsZip64 := writer.forceZip64
	if !archiveIsZip64 {
		for _, fi := range writer.fileInfoEntries {
			if fi.accumulatedSize >= zip64MagicVal || fi.headerOffset >= zip64MagicVal {
				archiveIsZip64 = true
				break
			}
		}
		if cdStartActualOffset >= zip64MagicVal || len(writer.fileInfoEntries) >= zip64EntriesMagicVal {
			archiveIsZip64 = true
		}
	}


	if archiveIsZip64 {
		// Write Zip64 EOCD Record
		zip64EOCD := Zip64EndOfCDRecord{
			Signature:                        zip64EndOfCDSignature,
			RecordSize:                       zip64EndOfCDRecordSize - 12, // Size of record after this field
			VersionMadeBy:                    zipVersion,
			VersionToExtract:                 zipVersion,
			DiskNumber:                       0,
			StartDiskNumber:                  0,
			NumberOfCDRecordEntries:          uint64(len(writer.fileInfoEntries)),
			TotalCDRecordEntries:             uint64(len(writer.fileInfoEntries)),
			CentralDirectorySize:             cdSize,
			StartingDiskCentralDirectoryOffset: cdStartActualOffset,
		}
		buf := new(bytes.Buffer)
		if err := binary.Write(buf, binary.LittleEndian, zip64EOCD); err != nil {
			return fmt.Errorf("binary.Write Zip64EndOfCDRecord failed: %w", err)
		}
		if err := writer.writeData(buf.Bytes()); err != nil { // This writeData advances the *global* totalBytes
			return fmt.Errorf("writer.writeData for Zip64EndOfCDRecord failed: %w", err)
		}
		offsetOfZip64EOCD := writer.lastOffsetCDFileHeader // Zip64 EOCD record is written after CD
		writer.lastOffsetCDFileHeader += uint64(zip64EndOfCDRecordSize) // Update offset to point after Zip64 EOCD

		// Write Zip64 EOCD Locator
		zip64Locator := Zip64EndOfCDRecordLocator{
			Signature:       zip64EndOfCDLocatorSignature,
			CDStartDiskNumber: 0,
			CDOffset:        offsetOfZip64EOCD, // Offset of the Zip64 EOCD Record
			NumberOfDisks:   1,
		}
		buf.Reset()
		if err := binary.Write(buf, binary.LittleEndian, zip64Locator); err != nil {
			return fmt.Errorf("binary.Write Zip64EndOfCDRecordLocator failed: %w", err)
		}
		if err := writer.writeData(buf.Bytes()); err != nil {
			return fmt.Errorf("writer.writeData for Zip64EndOfCDRecordLocator failed: %w", err)
		}
		writer.lastOffsetCDFileHeader += uint64(zip64EndOfCDLocatorSize) // Update offset to point after Zip64 EOCD Locator
	}

	// Write standard EndOfCDRecord
	eocd := EndOfCDRecord{
		Signature:             endOfCentralDirectorySignature,
		DiskNumber:            0,
		StartDiskNumber:       0,
		CommentLength:         0,
	}
	if archiveIsZip64 || len(writer.fileInfoEntries) >= zip64EntriesMagicVal {
		eocd.NumberOfCDRecordEntries = zip64EntriesMagicVal
		eocd.TotalCDRecordEntries = zip64EntriesMagicVal
	} else {
		eocd.NumberOfCDRecordEntries = uint16(len(writer.fileInfoEntries))
		eocd.TotalCDRecordEntries = uint16(len(writer.fileInfoEntries))
	}

	if archiveIsZip64 || cdSize >= zip64MagicVal {
		eocd.SizeOfCentralDirectory = zip64MagicVal
	} else {
		eocd.SizeOfCentralDirectory = uint32(cdSize)
	}

	if archiveIsZip64 || cdStartActualOffset >= zip64MagicVal {
		eocd.CentralDirectoryOffset = zip64MagicVal
	} else {
		eocd.CentralDirectoryOffset = uint32(cdStartActualOffset)
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, eocd); err != nil {
		return fmt.Errorf("binary.Write EndOfCDRecord failed: %w", err)
	}
	if err := writer.writeData(buf.Bytes()); err != nil {
		return fmt.Errorf("writer.writeData for EndOfCDRecord failed: %w", err)
	}
	// writer.lastOffsetCDFileHeader is already past Zip64 structures if they were written.
	// This standard EOCD is the last thing.

	return nil
}
