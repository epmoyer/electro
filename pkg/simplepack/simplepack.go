package simplepack

import (
	"app/pkg/quicklog"
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	minify "github.com/tdewolff/minify/v2"
	js "github.com/tdewolff/minify/v2/js"
)

const Version = "2.0.0"

const pathBuild = "project_build"
const pathEventData = pathBuild + "/app/js/event_data.js"
const pathTemp = pathBuild + "/temp"

type ReplacementDescriptorT struct {
	ContentType    string
	Regex          string
	ReplacementTag string
}

var replacementDescriptors = []ReplacementDescriptorT{
	{
		ContentType:    "javascript",
		Regex:          `^\s*<script\s*type="text/javascript"\s*src="(.*)"\s*>\s*</script>\s*$`,
		ReplacementTag: "script",
	},
	{
		ContentType:    "CSS",
		Regex:          `^\s*<link\s*rel="stylesheet"\s*type="text/css"\s*href="(.*)"\s*>\s*$`,
		ReplacementTag: "style",
	},
}

var qlog *quicklog.LoggerT = nil // Assigned at runtime

func init() {
	qlog = quicklog.GetLogger("default")
}

func Pack(pathFileIn string, pathFileOut string, enableMinify bool) error {

	qlog.Infof("SimplePack version: %q", Version)
	qlog.Infof("    Input file: %q", pathFileIn)
	qlog.Infof("    Output file: %q", pathFileOut)
	qlog.Infof("    Minify: %v", enableMinify)

	lines, err := readFileLines(pathFileIn)
	if err != nil {
		return fmt.Errorf("error reading input file %q as lines: %w", pathFileIn, err)
	}
	qlog.Debugf("Read %d lines", len(lines))

	newLines := []string{}
	for _, line := range lines {
		pathFileInParentDir := filepath.Dir(pathFileIn)
		expandedLines, err := expandLine(line, pathFileInParentDir, enableMinify)
		if err != nil {
			return fmt.Errorf("error expanding line: %w", err)
		}
		for _, expandedLine := range expandedLines {
			newLines = append(newLines, expandedLine)
		}
	}
	qlog.Infof("Writing: %q", pathFileOut)
	err = os.WriteFile(pathFileOut, []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("error writing output file: %w", err)
	}
	return nil
}

func expandLine(line string, pathFileInParentDir string, enableMinify bool) ([]string, error) {
	for _, descriptor := range replacementDescriptors {
		reg := regexp.MustCompile(descriptor.Regex)

		matches := reg.FindStringSubmatch(line)
		if len(matches) > 1 {
			pathFileToEmbed := matches[1]
			if strings.HasPrefix(pathFileToEmbed, "http://") || strings.HasPrefix(pathFileToEmbed, "https://") {
				// Do not embed remote files
				qlog.Infof("Skipping embedding remote %s file: %q", descriptor.ContentType, pathFileToEmbed)
				return []string{line}, nil
			}
			qlog.Infof("Embedding %s file: %q", descriptor.ContentType, pathFileToEmbed)
			// Read file to embed
			pathFileToEmbedFull := path.Join(pathFileInParentDir, pathFileToEmbed)
			data, err := os.ReadFile(pathFileToEmbedFull)
			if err != nil {
				return nil, fmt.Errorf("error reading file to embed (%q): %w", pathFileToEmbedFull, err)
			}
			content := string(data)
			if enableMinify {
				content, err = minifyContent(content, descriptor.ContentType)
				if err != nil {
					return nil, fmt.Errorf(
						"error minifying embedded %s file (%q): %w",
						descriptor.ContentType,
						pathFileToEmbedFull,
						err)
				}
			}
			// FIXME: Consider dropping indentation when minifying
			replacementTagStart := fmt.Sprintf("    <%s>", descriptor.ReplacementTag)
			replacementTagEnd := fmt.Sprintf("    </%s>", descriptor.ReplacementTag)
			embeddedLines := []string{replacementTagStart}
			for _, contentLine := range strings.Split(content, "\n") {
				embeddedLines = append(embeddedLines, "        "+contentLine)
			}
			embeddedLines = append(embeddedLines, replacementTagEnd)
			return embeddedLines, nil
		}
	}
	// No matches, return original line
	return []string{line}, nil
}

func minifyContent(content string, contentType string) (string, error) {
	if contentType != "javascript" {
		// FIXME: Implement CSS?
		return content, nil
	}
	m := minify.New()
	m.AddFunc("text/javascript", js.Minify)

	output, err := m.String("text/javascript", content)
	if err != nil {
		return "", err
	}
	return output, nil
}

// readFileLines reads data from a text file and returns a slice of file lines
func readFileLines(pathFile string) ([]string, error) {
	file, err := os.Open(pathFile)
	if err != nil {
		return []string{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// We set a large token sixe because we will be parsing HTML with
	// large (typically ~60K) inlined base64 data which appears on a
	// single line.
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return []string{}, err
	}
	return lines, nil
}
