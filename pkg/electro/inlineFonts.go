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
func makeHTMLFontsInline(inFilepath, outFilepath string) error {
	// Get the base directory of the input file
	basepath := filepath.Dir(strings.TrimRight(inFilepath, string(os.PathSeparator)))

	woffRe := regexp.MustCompile(`.*format\(["\']woff["\']\)`)
	woff2Re := regexp.MustCompile(`.*format\(["\']woff2["\']\)`)

	buf, err := os.ReadFile(inFilepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	lines := strings.Split(string(buf), "\n")
	linesOut := []string{}

	for _, line := range lines {
		// FIXME: Implement
		if woffRe.MatchString(line) {
			fmt.Printf("🟣 WOFF:%s\n", line)
			line = convertFont(basepath, line, "woff")
			fmt.Printf("    Converted: %s\n", line)
		}
		if woff2Re.MatchString(line) {
			fmt.Printf("🟣 WOFF2:%s\n", line)
			line = convertFont(basepath, line, "woff2")
			fmt.Printf("    Converted: %s\n", line)
		}
		linesOut = append(linesOut, line)
	}

	out := strings.Join(linesOut, "\n")

	err = os.WriteFile(outFilepath, []byte(out), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func convertFont(basepath, line, format string) string {
	urlRe := regexp.MustCompile(`url\(["\'](.*?)["\']\)`)
	urlExpressionRe := regexp.MustCompile(`rl\(["\'].*?["\']\)`)
	url := urlRe.FindAllString(line, -1)
	urlExpression := urlExpressionRe.FindAllString(line, -1)
	if len(url) != 1 || len(urlExpression) != 1 {
		fmt.Printf(
			"🔴  Expected to find 1 and only 1 font URL entry on line: %q."+
				"  Skipping font inline conversion.", line)
		return line
	}
	return line
}
