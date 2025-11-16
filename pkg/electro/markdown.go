package electro

import "fmt"

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

	return md, nil
}
