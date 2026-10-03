package transfer

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/realibrahimsql/Gocat/internal/logger"
)

// FileTransfer handles file transfer operations
type FileTransfer struct {
	ShowProgress   bool
	VerifyChecksum bool
	BufferSize     int
	RateLimit      int64 // bytes per second
}

// TransferInfo contains information about a file transfer
type TransferInfo struct {
	Filename string
	Size     int64
	MD5      string
	SHA256   string
}

// NewFileTransfer creates a new file transfer instance
func NewFileTransfer() *FileTransfer {
	return &FileTransfer{
		ShowProgress:   true,
		VerifyChecksum: true,
		BufferSize:     32768, // 32KB default
		RateLimit:      0,     // No limit by default
	}
}

// SendFile sends a file over the connection
func (ft *FileTransfer) SendFile(writer io.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %v", err)
	}

	// Send file info header
	info := TransferInfo{
		Filename: filepath.Base(filename),
		Size:     stat.Size(),
	}

	if ft.VerifyChecksum {
		info.MD5, info.SHA256, err = ft.calculateChecksums(filename)
		if err != nil {
			logger.Warn("Failed to calculate checksums: %v", err)
		}
	}

	// Send header
	header := fmt.Sprintf("GOCAT_FILE:%s:%d:%s:%s\n",
		info.Filename, info.Size, info.MD5, info.SHA256)
	if _, err := writer.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to send header: %v", err)
	}

	logger.Info("Sending file: %s (%d bytes)", info.Filename, info.Size)

	// Send file content with progress and rate limiting
	return ft.copyWithProgress(writer, file, info.Size, info.Filename)
}

// ReceiveFile receives a file from the connection.
// Header is read line-delimited with bufio so a coalesced header+data
// segment does not discard the leading file bytes, and short/fragmented
// headers are reassembled correctly.
func (ft *FileTransfer) ReceiveFile(reader io.Reader, outputDir string) error {
	br, ok := reader.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(reader)
	}
	header, err := br.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read header: %v", err)
	}
	if len(header) > 64*1024 {
		return fmt.Errorf("header too large: %d bytes", len(header))
	}

	info, err := ft.parseHeader(header)
	if err != nil {
		return fmt.Errorf("failed to parse header: %v", err)
	}
	if strings.Contains(info.Filename, "/") || strings.Contains(info.Filename, "\\") || info.Filename == "" || info.Filename == "." || info.Filename == ".." {
		// Prevent directory traversal; use base name only.
		info.Filename = filepath.Base(info.Filename)
	}
	if info.Size < 0 || info.Size > 100*1024*1024*1024 {
		return fmt.Errorf("invalid file size: %d", info.Size)
	}

	// Create output file
	outputPath := filepath.Join(outputDir, info.Filename)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	logger.Info("Receiving file: %s (%d bytes)", info.Filename, info.Size)

	// Receive exactly info.Size bytes; br still holds any over-read data.
	err = ft.copyWithProgress(file, io.LimitReader(br, info.Size), info.Size, info.Filename)
	if err != nil {
		return fmt.Errorf("failed to receive file: %v", err)
	}

	// Verify checksums if available
	if ft.VerifyChecksum && (info.MD5 != "" || info.SHA256 != "") {
		return ft.verifyFile(outputPath, info)
	}

	return nil
}

// copyWithProgress copies data with progress indication and rate limiting
func (ft *FileTransfer) copyWithProgress(dst io.Writer, src io.Reader, size int64, filename string) error {
	buffer := make([]byte, ft.BufferSize)
	var written int64
	startTime := time.Now()

	for {
		// Rate limiting
		if ft.RateLimit > 0 {
			elapsed := time.Since(startTime)
			expectedTime := time.Duration(written * int64(time.Second) / ft.RateLimit)
			if elapsed < expectedTime {
				time.Sleep(expectedTime - elapsed)
			}
		}

		n, err := src.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err := writeFullTransfer(dst, buffer[:n]); err != nil {
			return err
		}

		written += int64(n)

		// Show progress
		if ft.ShowProgress && size > 0 {
			ft.showProgress(filename, written, size)
		}
		if size > 0 && written > size {
			return fmt.Errorf("peer sent %d bytes, more than announced %d", written, size)
		}
	}

	if size > 0 && written != size {
		return fmt.Errorf("short transfer: got %d bytes, expected %d", written, size)
	}

	if ft.ShowProgress {
		fmt.Printf("\n")
		logger.Info("Transfer completed: %s (%d bytes in %v)",
			filename, written, time.Since(startTime))
	}

	return nil
}

func writeFullTransfer(dst io.Writer, p []byte) error {
	for len(p) > 0 {
		n, err := dst.Write(p)
		if err != nil {
			return err
		}
		if n < 0 || n > len(p) {
			return fmt.Errorf("invalid write count %d for %d bytes", n, len(p))
		}
		p = p[n:]
	}
	return nil
}

// showProgress displays transfer progress
func (ft *FileTransfer) showProgress(filename string, current, total int64) {
	if total == 0 {
		return
	}

	percent := float64(current) / float64(total) * 100
	barWidth := 50
	filled := int(percent / 100 * float64(barWidth))

	bar := ""
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	fmt.Printf("\r%s [%s] %.1f%% (%d/%d bytes)",
		filename, bar, percent, current, total)
}

// calculateChecksums calculates MD5 and SHA256 checksums
func (ft *FileTransfer) calculateChecksums(filename string) (string, string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	md5Hash := md5.New()
	sha256Hash := sha256.New()

	writer := io.MultiWriter(md5Hash, sha256Hash)
	_, err = io.Copy(writer, file)
	if err != nil {
		return "", "", err
	}

	return fmt.Sprintf("%x", md5Hash.Sum(nil)),
		fmt.Sprintf("%x", sha256Hash.Sum(nil)), nil
}

// parseHeader parses the file transfer header.
// Format: GOCAT_FILE:filename:size:md5:sha256\n
// Filename may itself contain ':', so the last three fields are parsed
// from the end (size, md5, sha256) and the remainder is the filename.
func (ft *FileTransfer) parseHeader(header string) (*TransferInfo, error) {
	header = strings.TrimSuffix(header, "\n")
	header = strings.TrimSuffix(header, "\r")
	const prefix = "GOCAT_FILE:"
	if !strings.HasPrefix(header, prefix) {
		return nil, fmt.Errorf("invalid header format: missing GOCAT_FILE prefix")
	}
	rest := strings.TrimPrefix(header, prefix)
	parts := strings.Split(rest, ":")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid header format: expected filename:size:md5:sha256")
	}
	// Last three segments are size/md5/sha256; everything before is filename.
	sha := parts[len(parts)-1]
	md5sum := parts[len(parts)-2]
	sizeStr := parts[len(parts)-3]
	filename := strings.Join(parts[:len(parts)-3], ":")
	if filename == "" {
		return nil, fmt.Errorf("invalid header: empty filename")
	}

	info := &TransferInfo{
		Filename: filename,
	}

	if _, err := fmt.Sscanf(sizeStr, "%d", &info.Size); err != nil {
		return nil, fmt.Errorf("invalid size: %v", err)
	}
	if info.Size < 0 {
		return nil, fmt.Errorf("invalid size: %d", info.Size)
	}

	info.MD5 = md5sum
	info.SHA256 = sha

	return info, nil
}

// verifyFile verifies file checksums
func (ft *FileTransfer) verifyFile(filename string, expected *TransferInfo) error {
	md5Sum, sha256Sum, err := ft.calculateChecksums(filename)
	if err != nil {
		return fmt.Errorf("failed to calculate checksums: %v", err)
	}

	if expected.MD5 != "" && md5Sum != expected.MD5 {
		return fmt.Errorf("MD5 checksum mismatch: expected %s, got %s", expected.MD5, md5Sum)
	}

	if expected.SHA256 != "" && sha256Sum != expected.SHA256 {
		return fmt.Errorf("SHA256 checksum mismatch: expected %s, got %s", expected.SHA256, sha256Sum)
	}

	logger.Info("File verification successful: %s", filename)
	return nil
}
