package electro

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

// getMimetype determines MIME type by reading file content first,
// then falls back to file extension
func getMimetype(pathFile string) string {
	// First, try to detect from file content
	file, err := os.Open(pathFile)
	if err != nil {
		// If we can't open the file, fall back to extension only
		return mimetypeFromExtension(pathFile)
	}
	defer file.Close()

	// Read first 512 bytes for content detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		// If we can't read, fall back to extension
		return mimetypeFromExtension(pathFile)
	}

	// Detect content type from the buffer
	contentType := http.DetectContentType(buffer[:n])

	// http.DetectContentType returns "application/octet-stream"
	// when it can't determine the type, so fall back to extension
	if contentType == "application/octet-stream" {
		if extType := mimetypeFromExtension(pathFile); extType != "" {
			return extType
		}
	}

	return contentType
}

// mimetypeFromExtension gets MIME type based on file extension
func mimetypeFromExtension(pathFile string) string {
	ext := filepath.Ext(pathFile)
	mimeType := mime.TypeByExtension(ext)

	if mimeType == "" {
		return "application/octet-stream"
	}

	return mimeType
}
