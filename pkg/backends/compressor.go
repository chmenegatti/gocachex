package backends

import (
	"bytes"
	"compress/gzip"
	"io"
)

// GzipCompressor implements gzip compression.
type GzipCompressor struct{}

// Compress compresses data using gzip.
func (g *GzipCompressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decompress decompresses gzip data.
func (g *GzipCompressor) Decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// Algorithm returns the compression algorithm name.
func (g *GzipCompressor) Algorithm() string {
	return "gzip"
}

// LZ4Compressor implements LZ4 compression.
// Note: This is a placeholder implementation. In a real implementation,
// you would use a library like github.com/pierrec/lz4/v4
type LZ4Compressor struct{}

// Compress compresses data using LZ4 (placeholder).
func (l *LZ4Compressor) Compress(data []byte) ([]byte, error) {
	// For now, fall back to gzip
	// In a real implementation, use LZ4 library
	compressor := &GzipCompressor{}
	return compressor.Compress(data)
}

// Decompress decompresses LZ4 data (placeholder).
func (l *LZ4Compressor) Decompress(data []byte) ([]byte, error) {
	// For now, fall back to gzip
	// In a real implementation, use LZ4 library
	compressor := &GzipCompressor{}
	return compressor.Decompress(data)
}

// Algorithm returns the compression algorithm name.
func (l *LZ4Compressor) Algorithm() string {
	return "lz4"
}

// SnappyCompressor implements Snappy compression.
// Note: This is a placeholder implementation. In a real implementation,
// you would use a library like github.com/golang/snappy
type SnappyCompressor struct{}

// Compress compresses data using Snappy (placeholder).
func (s *SnappyCompressor) Compress(data []byte) ([]byte, error) {
	// For now, fall back to gzip
	// In a real implementation, use Snappy library
	compressor := &GzipCompressor{}
	return compressor.Compress(data)
}

// Decompress decompresses Snappy data (placeholder).
func (s *SnappyCompressor) Decompress(data []byte) ([]byte, error) {
	// For now, fall back to gzip
	// In a real implementation, use Snappy library
	compressor := &GzipCompressor{}
	return compressor.Decompress(data)
}

// Algorithm returns the compression algorithm name.
func (s *SnappyCompressor) Algorithm() string {
	return "snappy"
}

// CompressData is a helper function to compress data if a compressor is provided.
func CompressData(compressor Compressor, data []byte) ([]byte, error) {
	if compressor == nil {
		return data, nil
	}
	return compressor.Compress(data)
}

// DecompressData is a helper function to decompress data if a compressor is provided.
func DecompressData(compressor Compressor, data []byte) ([]byte, error) {
	if compressor == nil {
		return data, nil
	}
	return compressor.Decompress(data)
}

// ShouldCompress determines if data should be compressed based on size and type.
func ShouldCompress(data []byte, minSize int) bool {
	// Only compress if data is larger than minimum size
	if minSize > 0 && len(data) < minSize {
		return false
	}

	// Don't compress already compressed data (simple heuristic)
	if len(data) > 10 {
		// Check for common compressed file signatures
		if data[0] == 0x1f && data[1] == 0x8b { // gzip
			return false
		}
		if data[0] == 0x50 && data[1] == 0x4b { // zip
			return false
		}
		if len(data) > 3 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4e && data[3] == 0x47 { // PNG
			return false
		}
		if len(data) > 1 && data[0] == 0xff && data[1] == 0xd8 { // JPEG
			return false
		}
	}

	return true
}
