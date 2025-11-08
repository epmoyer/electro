package electro

import (
	"encoding/base64"
	"os"
)

func fileToBase64(pathFile string) (string, error) {
	// Read the entire file
	data, err := os.ReadFile(pathFile)
	if err != nil {
		return "", err
	}

	// Encode to base64
	encodedStr := base64.StdEncoding.EncodeToString(data)

	return encodedStr, nil
}
