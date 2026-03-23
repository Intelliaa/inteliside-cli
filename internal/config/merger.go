package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DeepMerge performs an additive-only merge of src into dst.
// - New keys are added
// - Existing scalars are NEVER overwritten
// - Objects are recursively merged
// - Arrays are unioned (no duplicates)
func DeepMerge(dst, src map[string]any) map[string]any {
	if dst == nil {
		dst = make(map[string]any)
	}
	for key, srcVal := range src {
		dstVal, exists := dst[key]
		if !exists {
			dst[key] = srcVal
			continue
		}

		switch sv := srcVal.(type) {
		case map[string]any:
			if dv, ok := dstVal.(map[string]any); ok {
				dst[key] = DeepMerge(dv, sv)
			}
			// types differ: keep existing
		case []any:
			if dv, ok := dstVal.([]any); ok {
				dst[key] = unionArrays(dv, sv)
			}
			// types differ: keep existing
		default:
			// scalar: keep existing, never overwrite
		}
	}
	return dst
}

// unionArrays merges two arrays, skipping duplicates.
func unionArrays(dst, src []any) []any {
	seen := make(map[string]bool)
	for _, v := range dst {
		seen[toKey(v)] = true
	}
	for _, v := range src {
		k := toKey(v)
		if !seen[k] {
			dst = append(dst, v)
			seen[k] = true
		}
	}
	return dst
}

func toKey(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// MergeJSONFile reads a JSON file, deep-merges incoming data, and writes back.
func MergeJSONFile(path string, incoming map[string]any) error {
	existing := make(map[string]any)

	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &existing)
	}

	merged := DeepMerge(existing, incoming)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	out, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, out, 0644)
}
