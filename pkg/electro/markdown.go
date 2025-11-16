package electro

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"
)

type mdRendererT struct {
	Markdown           string
	Substitutions      map[string]string
	DoStripFrontmatter bool
}

func newMdRenderer(markdown string) *mdRendererT {
	return &mdRendererT{
		Markdown:           markdown,
		Substitutions:      make(map[string]string),
		DoStripFrontmatter: false,
	}
}

func (r *mdRendererT) Render() (string, error) {
	md := r.Markdown

	// -------------------------
	// Pre-parser
	// -------------------------
	md, err := r.PreParseMarkdown(md)
	if err != nil {
		return "", fmt.Errorf("error pre-parsing markdown content: %w", err)
	}

	// -------------------------
	// Render markdown to HTML
	// -------------------------
	var bufHtmlBytes bytes.Buffer
	mdConverter := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
			),
		),
		goldmark.WithRendererOptions(
			gmhtml.WithUnsafe(),
		),
	)
	err = mdConverter.Convert([]byte(md), &bufHtmlBytes)
	if err != nil {
		return "", fmt.Errorf("error converting markdown to HTML: %w", err)
	}
	html := bufHtmlBytes.String()

	// -------------------------
	// Post-parser
	// -------------------------
	html, err = r.PostParseHtml(html)
	if err != nil {
		return "", fmt.Errorf("error post-processing HTML: %w", err)
	}

	return html, nil
}

func (r *mdRendererT) PreParseMarkdown(md string) (string, error) {
	var err error

	// -------------------------
	// Strip frontmatter
	// -------------------------
	if r.DoStripFrontmatter {
		md, err = r.stripFrontmatter(md)
		if err != nil {
			return "", fmt.Errorf("error stripping frontmatter: %w", err)
		}
	}

	// FIXME:md:finish implementation

	return md, nil
}

func (r *mdRendererT) stripFrontmatter(md string) (string, error) {
	lines := strings.Split(md, "\n")

	// Frontmatter must start on the first line
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return md, nil // No frontmatter
	}

	// Find closing ---
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			// Found end of frontmatter, return everything after
			if i+1 < len(lines) {
				return strings.Join(lines[i+1:], "\n"), nil
			}
			return "", nil // File was only frontmatter
		}
	}

	return md, fmt.Errorf("frontmatter start found but no closing '---'")
}

func (r *mdRendererT) PostParseHtml(html string) (string, error) {
	for placeholder, final := range r.Substitutions {
		html = strings.ReplaceAll(html, placeholder, final)
	}
	return html, nil
}
