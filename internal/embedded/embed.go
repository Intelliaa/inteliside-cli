package embedded

import (
	"embed"
	"regexp"
	"strings"
)

//go:embed rules/*.md
var rulesFS embed.FS

//go:embed templates/*.md
var templatesFS embed.FS

// RuleFiles returns a map[filename]content for all embedded rule files.
func RuleFiles() (map[string]string, error) {
	entries, err := rulesFS.ReadDir("rules")
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := rulesFS.ReadFile("rules/" + e.Name())
		if err != nil {
			return nil, err
		}
		result[e.Name()] = string(data)
	}
	return result, nil
}

// RuleFileNames returns the list of embedded rule file names.
func RuleFileNames() []string {
	entries, err := rulesFS.ReadDir("rules")
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

// ATLTemplate returns the embedded CLAUDE.md template for ATL projects.
func ATLTemplate() string {
	data, err := templatesFS.ReadFile("templates/CLAUDE-atl.md")
	if err != nil {
		return ""
	}
	return string(data)
}

// LegacyTemplate returns the embedded CLAUDE.md template for legacy context.
func LegacyTemplate() string {
	data, err := templatesFS.ReadFile("templates/CLAUDE-legacy.md")
	if err != nil {
		return ""
	}
	return string(data)
}

var placeholderRe = regexp.MustCompile(`\{\{(\w+)\}\}`)

// RenderTemplate replaces {{key}} placeholders with values from vars.
// Unresolved placeholders become <!-- TODO: key --> for easy detection.
func RenderTemplate(tmpl string, vars map[string]string) string {
	result := tmpl
	for k, v := range vars {
		if v != "" {
			result = strings.ReplaceAll(result, "{{"+k+"}}", v)
		}
	}
	// Convert remaining placeholders to TODO markers
	result = placeholderRe.ReplaceAllString(result, "<!-- TODO: $1 -->")
	return result
}

// CountTODOs returns the number of <!-- TODO: ... --> markers in content.
func CountTODOs(content string) int {
	return strings.Count(content, "<!-- TODO:")
}
