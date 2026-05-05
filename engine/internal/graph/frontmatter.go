package graph

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// fmFieldKeys maps human-readable markdown field names to YAML frontmatter keys.
var fmFieldKeys = map[string]string{
	"Realised By":    "realised_by",
	"Realises":       "realises",
	"Implemented By": "implemented_by",
	"Refines":        "refines",
	"Requirements":   "requirements",
	"Traces":         "traces",
}

// splitFrontmatter separates the YAML frontmatter from the markdown body.
// The file must start with "---\n" and the frontmatter must end with "\n---\n".
// Returns the raw YAML bytes (without delimiters), the body text, and true.
// Returns (nil, "", false) when no frontmatter is present.
func splitFrontmatter(data []byte) (fmBytes []byte, body string, ok bool) {
	s := string(data)
	if !strings.HasPrefix(s, "---\n") {
		return nil, "", false
	}
	rest := s[4:]
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		return nil, "", false
	}
	// fmBytes contains the YAML content with its trailing newline.
	return []byte(rest[:idx+1]), rest[idx+5:], true
}

// frontmatterStringList returns the list of strings stored under key in the
// raw YAML frontmatter bytes. Missing or null values return (nil, nil).
func frontmatterStringList(fmBytes []byte, key string) ([]string, error) {
	var m map[string]interface{}
	if err := yaml.Unmarshal(fmBytes, &m); err != nil {
		return nil, err
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil, nil
	}
	seq, ok := v.([]interface{})
	if !ok {
		return nil, nil
	}
	out := make([]string, 0, len(seq))
	for _, item := range seq {
		if s, ok := item.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out, nil
}

// fmUpdate describes a single key update to apply to a YAML frontmatter block.
// When isScalar is true, Scalar is used as a YAML value literal; otherwise
// Paths is used as a YAML sequence.
// When omitIfAbsent is true, the key is only written if it already exists in
// the frontmatter or there are actual paths to write.
type fmUpdate struct {
	key          string
	paths        []string
	scalar       string
	isScalar     bool
	omitIfAbsent bool
}

// applyFMUpdates returns new file content with the given frontmatter keys updated.
// It returns an error if the file has no frontmatter.
func applyFMUpdates(content string, updates []fmUpdate) (string, error) {
	fmBytes, body, ok := splitFrontmatter([]byte(content))
	if !ok {
		return "", fmt.Errorf("no frontmatter found")
	}
	fm := string(fmBytes)
	for _, u := range updates {
		if u.isScalar {
			fm = setFMScalar(fm, u.key, u.scalar)
		} else if u.omitIfAbsent && len(u.paths) == 0 && !fmKeyExists(fm, u.key) {
			continue
		} else {
			fm = setFMSequence(fm, u.key, u.paths)
		}
	}
	return "---\n" + fm + "---\n" + body, nil
}

// setFMSequence replaces or appends the YAML sequence for key in the raw
// frontmatter string fm. Items are written with 4-space indentation.
// An empty or nil paths slice writes the inline form "key: []".
func setFMSequence(fm, key string, paths []string) string {
	var newBlock string
	if len(paths) == 0 {
		newBlock = key + ": []\n"
	} else {
		var sb strings.Builder
		sb.WriteString(key + ":\n")
		for _, p := range paths {
			sb.WriteString("    - " + p + "\n")
		}
		newBlock = sb.String()
	}
	return replaceFMKey(fm, key, newBlock)
}

// setFMScalar replaces or appends a YAML scalar key in the raw frontmatter string fm.
// value should be a valid YAML scalar literal (e.g. "~", a plain string, or a
// quoted string).
func setFMScalar(fm, key, value string) string {
	return replaceFMKey(fm, key, key+": "+value+"\n")
}

// fmKeyExists reports whether key appears as a top-level entry in the raw
// frontmatter string fm.
func fmKeyExists(fm, key string) bool {
	for _, line := range strings.Split(fm, "\n") {
		t := strings.TrimRight(line, " \t")
		if t == key+":" || strings.HasPrefix(t, key+": ") {
			return true
		}
	}
	return false
}

// replaceFMKey replaces the existing key block (key line + indented continuation)
// in the raw frontmatter string fm with newBlock. If the key is absent, newBlock
// is appended.
func replaceFMKey(fm, key, newBlock string) string {
	lines := strings.Split(fm, "\n")
	startIdx := -1
	for i, line := range lines {
		t := strings.TrimRight(line, " \t")
		if t == key+":" || strings.HasPrefix(t, key+": ") {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		return strings.TrimRight(fm, "\n") + "\n" + newBlock
	}
	// Find the end of the old value block (indented continuation lines).
	endIdx := startIdx + 1
	for endIdx < len(lines) {
		if strings.HasPrefix(lines[endIdx], " ") || strings.HasPrefix(lines[endIdx], "\t") {
			endIdx++
		} else {
			break
		}
	}
	newBlockLines := strings.Split(strings.TrimRight(newBlock, "\n"), "\n")
	result := make([]string, 0, len(lines))
	result = append(result, lines[:startIdx]...)
	result = append(result, newBlockLines...)
	result = append(result, lines[endIdx:]...)
	return strings.Join(result, "\n")
}
