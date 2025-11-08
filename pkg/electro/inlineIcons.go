package electro

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// makeHTMLIconsInline takes an HTML file and writes a new version with
// inline Base64 encoded fonts
func makeHTMLIconsInline(pathFileIn, pathFileOut string) error {
	var err error
	var lines []string

	// Get the base directory of the input file
	basepath := filepath.Dir(strings.TrimRight(pathFileIn, string(os.PathSeparator)))

	iconRe := regexp.MustCompile(`.*rel=["\']shortcut icon["\']`)

	lines, err = readFileLines(pathFileIn)
	if err != nil {
		return fmt.Errorf("failed to read lines of file %q: %w", pathFileIn, err)
	}

	linesOut := []string{}

	for _, line := range lines {
		if iconRe.MatchString(line) {
			fmt.Printf("🟣 ICON:%s\n", line)
			line, err = convertIcon(basepath, line)
			if err != nil {
				return fmt.Errorf("failed to inline icon: %w", err)
			}
			fmt.Printf("    Converted: %s\n", line)
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

func convertIcon(basepath, line string) (string, error) {
	return line, nil
}
