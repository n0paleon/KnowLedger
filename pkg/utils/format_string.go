package utils

import "strings"

func FormatTagsStrToSlice(tags string) []string {
	if tags == "" {
		return []string{}
	}

	tagSlice := strings.Split(tags, ",")
	seen := make(map[string]bool)
	result := make([]string, 0, len(tagSlice))

	for _, tag := range tagSlice {
		trimmedTag := strings.TrimSpace(tag)
		if trimmedTag != "" && !seen[trimmedTag] {
			seen[trimmedTag] = true
			result = append(result, trimmedTag)
		}
	}

	return result
}

func NormalizeTagName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
