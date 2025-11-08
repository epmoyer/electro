package electro

import (
	"fmt"
	"os"
	"strings"
)

// makeHTMLFontsInline takes an HTML file and writes a new version with
// inline Base64 encoded fonts
func makeHTMLFontsInline(inFilepath, outFilepath string) error {
	// Get the base directory of the input file
	// basepath := filepath.Dir(strings.TrimRight(inFilepath, string(os.PathSeparator)))

	buf, err := os.ReadFile(inFilepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	lines := strings.Split(string(buf), "\n")
	linesOut := []string{}

	for _, line := range lines {
		// FIXME: Implement
		linesOut = append(linesOut, line)
	}

	out := strings.Join(linesOut, "\n")

	err = os.WriteFile(outFilepath, []byte(out), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
