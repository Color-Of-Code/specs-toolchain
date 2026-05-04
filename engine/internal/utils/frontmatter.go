package utils

import (
	"bytes"
)

// ExtractFrontmatter extracts the YAML frontmatter from the given data.
// It returns the frontmatter (if present), the remaining content, and the number of newlines consumed.
func ExtractFrontmatter(data []byte) (frontmatter, body []byte, lineOffset int) {
	if !bytes.HasPrefix(data, []byte("---\n")) {
		return nil, data, 0
	}
	idx := bytes.Index(data[4:], []byte("\n---\n"))
	if idx < 0 {
		return nil, data, 0
	}
	frontmatter = data[:4+idx+5] // "---\n" + yaml + "\n---\n"
	body = data[4+idx+5:]        // body after closing delimiter
	lineOffset = bytes.Count(frontmatter, []byte("\n"))
	return frontmatter, body, lineOffset
}
