package electro

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// makeHTMLImagesInline takes an HTML file and writes a new version with
// inline Base64 encoded images
func makeHTMLImagesInline(inFilepath, outFilepath string) error {
	// Get the base directory of the input file
	basepath := filepath.Dir(strings.TrimRight(inFilepath, string(os.PathSeparator)))

	// Open and parse the HTML file
	file, err := os.Open(inFilepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find all img tags and process them
	doc.Find("img").Each(func(i int, img *goquery.Selection) {
		sourcePath, exists := img.Attr("src")
		if !exists {
			return
		}

		// Unescape URL-encoded characters (e.g., %20 for space)
		sourcePath, err := url.PathUnescape(sourcePath)
		if err != nil {
			// If unescaping fails, use original path
			sourcePath, _ = img.Attr("src")
		}

		var imgPath string
		if strings.HasPrefix(sourcePath, "/") {
			// Absolute path
			imgPath = sourcePath
		} else {
			// Relative path
			imgPath = path.Join(basepath, sourcePath)
		}

		// Get MIME type and encode to base64
		mimetype := getMimetype(imgPath)
		base64Data, err := fileToBase64(imgPath)
		if err != nil {
			// Skip this image if we can't read it
			return
		}

		// Set the data URI
		dataURI := fmt.Sprintf("data:%s;base64,%s", mimetype, base64Data)
		img.SetAttr("src", dataURI)
	})

	// Write the modified HTML to output file
	html, err := doc.Html()
	if err != nil {
		return fmt.Errorf("failed to generate HTML: %w", err)
	}

	err = os.WriteFile(outFilepath, []byte(html), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
