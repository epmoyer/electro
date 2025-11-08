package electro

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// makeHTMLFontsInline takes an HTML file and writes a new version with
// inline Base64 encoded fonts
func makeHTMLFontsInline(pathFileIn, pathFileOut string) error {
	var err error
	var lines []string

	// Get the base directory of the input file
	basepath := filepath.Dir(strings.TrimRight(pathFileIn, string(os.PathSeparator)))

	woffRe := regexp.MustCompile(`.*format\(["\']woff["\']\)`)
	woff2Re := regexp.MustCompile(`.*format\(["\']woff2["\']\)`)

	lines, err = readFileLines(pathFileIn)
	if err != nil {
		return fmt.Errorf("failed to read lines of file %q: %w", pathFileIn, err)
	}

	linesOut := []string{}

	for _, line := range lines {
		if woffRe.MatchString(line) {
			// fmt.Printf("🟣 WOFF:%s\n", line)
			line, err = convertFont(basepath, line, "woff")
			if err != nil {
				return fmt.Errorf("failed to inline woff font: %w", err)
			}
			// fmt.Printf("    Converted: %s\n", line)
		}
		if woff2Re.MatchString(line) {
			// fmt.Printf("🟣 WOFF2:%s\n", line)
			line, err = convertFont(basepath, line, "woff2")
			if err != nil {
				return fmt.Errorf("failed to inline woff2 font: %w", err)
			}
			// fmt.Printf("    Converted: %s\n", line)
		}
		linesOut = append(linesOut, line)
	}

	out := strings.Join(linesOut, "\n")

	err = os.WriteFile(pathFileOut, []byte(out), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func convertFont(basepath, line, format string) (string, error) {
	urlRe := regexp.MustCompile(`url\(["\'](.*?)["\']\)`)
	urlExpressionRe := regexp.MustCompile(`url\(["\'].*?["\']\)`)
	urls := urlRe.FindAllStringSubmatch(line, -1)
	urlExpressions := urlExpressionRe.FindAllString(line, -1)
	if len(urls) != 1 || len(urlExpressions) != 1 {
		return "", fmt.Errorf("expected to find 1 and only 1 font URL entry on line: %q", line)
	}
	url := urls[0][1]
	urlExpression := urlExpressions[0]
	// fmt.Printf("🟣 %q :: %q\n", url, urlExpression)

	url = strings.TrimPrefix(url, "/")
	pathFont := basepath + "/" + url
	if !pathExists(pathFont) || !pathIsFile(pathFont) {
		return "", fmt.Errorf("no font file found at url: %q", pathFont)
	}
	fontBase64, err := fileToBase64(pathFont)
	if err != nil {
		return "", err
	}
	urlInline := fmt.Sprintf("url(data:application/x-font-%s;charset=utf-8;base64,%s)",
		format, fontBase64)
	line = strings.Replace(line, urlExpression, urlInline, 1)

	return line, nil
}
