package resource

import "strings"

// NameFromPath prefers the concrete admin resource segment and falls back to a
// framework route parameter when the request path still contains a template.
func NameFromPath(path, routeParam string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for index, part := range parts {
		if part != "admin" || index+1 >= len(parts) {
			continue
		}
		if candidate := concreteResourceName(parts[index+1]); candidate != "" {
			return candidate
		}
		break
	}
	return concreteResourceName(routeParam)
}

func concreteResourceName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "{}:/") {
		return ""
	}
	return value
}
