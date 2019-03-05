package format

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func indent(in string) string {
	parts := strings.Split(in, "\n")

	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = "  " + part
		}
	}

	return strings.TrimLeft(strings.Join(parts, "\n"), " ")
}

func sortKeys(data map[string]interface{}, path []string) []string {
	// See if we have some required-order keys
	order := ordering(path)

	done := make(map[string]bool)
	orderedKeys := make([]string, 0, len(data))
	otherKeys := make([]string, 0, len(data))

	// Apply required keys
	for _, orderedKey := range order {
		if _, ok := data[orderedKey]; ok {
			orderedKeys = append(orderedKeys, orderedKey)
			done[orderedKey] = true
		}
	}

	// Now the remainder of the keys
	for key, _ := range data {
		if !done[key] {
			otherKeys = append(otherKeys, key)
		}
	}
	sort.Strings(otherKeys)

	return append(orderedKeys, otherKeys...)
}

func intrinsicKey(data map[string]interface{}) (string, bool) {
	if len(data) != 1 {
		return "", false
	}

	// We know there's one key
	key := reflect.ValueOf(data).MapKeys()[0].String()
	if key == "Ref" || strings.HasPrefix(key, "Fn::") {
		return key, true
	}

	return "", false
}

func formatIntrinsic(key string, data interface{}, path []string) string {
	shortKey := strings.Replace(key, "Fn::", "", 1)

	fmtValue := yaml(data, path)

	switch data.(type) {
	case map[string]interface{}:
		return fmt.Sprintf("!%s\n  %s", shortKey, indent(fmtValue))
	case []interface{}:
		return fmt.Sprintf("!%s\n  %s", shortKey, indent(fmtValue))
	default:
		return fmt.Sprintf("!%s %s", shortKey, yaml(data, path))
	}
}

func formatMap(data map[string]interface{}, path []string) string {
	if len(data) == 0 {
		return "{}"
	}

	keys := sortKeys(data, path)

	parts := make([]string, len(keys))

	for i, key := range keys {
		value := data[key]
		fmtValue := yaml(value, append(path, key))

		switch v := value.(type) {
		case map[string]interface{}:
			if iKey, ok := intrinsicKey(v); ok {
				fmtValue = formatIntrinsic(iKey, v[iKey], append(path, key))
				fmtValue = fmt.Sprintf("%s: %s", key, fmtValue)
			} else {
				fmtValue = fmt.Sprintf("%s:\n  %s", key, indent(fmtValue))
			}
		case []interface{}:
			fmtValue = fmt.Sprintf("%s:\n  %s", key, indent(fmtValue))
		default:
			fmtValue = fmt.Sprintf("%s: %s", key, fmtValue)
		}

		parts[i] = fmtValue
	}

	joiner := "\n"

	if len(path) <= 1 {
		joiner = "\n\n"
	}

	return strings.Join(parts, joiner)
}

func formatList(data []interface{}, path []string) string {
	if len(data) == 0 {
		return "[]"
	}

	parts := make([]string, len(data))

	for i, value := range data {
		fmtValue := yaml(value, append(path, string(i)))

		parts[i] = fmt.Sprintf("- %s", indent(fmtValue))
	}

	return strings.Join(parts, "\n")
}

func formatString(data string) string {
	quote := false

	switch {
	case data == "Yes" || data == "No":
		quote = true
	case strings.ContainsAny(string(data[0]), "0123456789!&*?,#|>@`\"'[{:"):
		quote = true
	case strings.ContainsAny(data, "\n"):
		quote = true
	}

	if quote {
		return fmt.Sprintf("%q", data)
	}

	return data
}

func yaml(data interface{}, path []string) string {
	switch value := data.(type) {
	case map[string]interface{}:
		return formatMap(value, path)
	case []interface{}:
		return formatList(value, path)
	case string:
		return formatString(value)
	default:
		return fmt.Sprint(value)
	}
}

func Yaml(data interface{}) string {
	return yaml(data, make([]string, 0))
}
