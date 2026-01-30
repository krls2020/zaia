package knowledge

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed embed/**/*.md
var contentFS embed.FS

// Document represents a parsed knowledge document.
type Document struct {
	Path        string   // embed/services/postgresql.md
	URI         string   // zerops://docs/services/postgresql
	Title       string   // PostgreSQL on Zerops
	Keywords    []string // [postgresql, postgres, sql, ...]
	TLDR        string   // One-sentence summary
	Content     string   // Full markdown content
	Description string   // TL;DR or first paragraph
}

// loadFromEmbedded walks the embedded filesystem and parses all markdown documents.
func loadFromEmbedded() map[string]*Document {
	docs := make(map[string]*Document)
	fs.WalkDir(contentFS, "embed", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, err := contentFS.ReadFile(path)
		if err != nil {
			return nil
		}
		doc := parseDocument(path, string(data))
		docs[doc.URI] = doc
		return nil
	})
	return docs
}

func parseDocument(path, content string) *Document {
	uri := pathToURI(path)
	title := extractTitle(content)
	keywords := extractKeywords(content)
	tldr := extractTLDR(content)

	desc := tldr
	if desc == "" {
		desc = extractFirstParagraph(content)
	}

	return &Document{
		Path:        path,
		URI:         uri,
		Title:       title,
		Keywords:    keywords,
		TLDR:        tldr,
		Content:     content,
		Description: desc,
	}
}

func pathToURI(fsPath string) string {
	// Remove "embed/" prefix
	rel := strings.TrimPrefix(fsPath, "embed/")
	// Remove .md extension
	rel = strings.TrimSuffix(rel, ".md")
	return "zerops://docs/" + rel
}

func uriToPath(uri string) string {
	rel := strings.TrimPrefix(uri, "zerops://docs/")
	return "embed/" + rel + ".md"
}

func extractTitle(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

func extractKeywords(content string) []string {
	lines := strings.Split(content, "\n")
	inKeywords := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## Keywords" {
			inKeywords = true
			continue
		}
		if inKeywords {
			if trimmed == "" || strings.HasPrefix(trimmed, "##") {
				break
			}
			parts := strings.Split(trimmed, ",")
			var keywords []string
			for _, p := range parts {
				kw := strings.TrimSpace(p)
				if kw != "" {
					keywords = append(keywords, strings.ToLower(kw))
				}
			}
			return keywords
		}
	}
	return nil
}

func extractTLDR(content string) string {
	lines := strings.Split(content, "\n")
	inTLDR := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## TL;DR" {
			inTLDR = true
			continue
		}
		if inTLDR {
			if trimmed == "" {
				continue
			}
			if strings.HasPrefix(trimmed, "##") {
				break
			}
			return trimmed
		}
	}
	return ""
}

func extractFirstParagraph(content string) string {
	lines := strings.Split(content, "\n")
	var para []string
	pastTitle := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			pastTitle = true
			continue
		}
		if !pastTitle {
			continue
		}
		if trimmed == "" && len(para) > 0 {
			break
		}
		if trimmed != "" && !strings.HasPrefix(trimmed, "##") {
			para = append(para, trimmed)
		}
	}
	result := strings.Join(para, " ")
	if len(result) > 200 {
		return result[:200] + "..."
	}
	return result
}
