package content

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

const (
	maxDocumentDepth = 32
	maxDocumentNodes = 10000
	maxTextLength    = 100000
)

var allowedNodeTypes = map[string]bool{
	"doc":            true,
	"paragraph":      true,
	"heading":        true,
	"bulletList":     true,
	"orderedList":    true,
	"listItem":       true,
	"blockquote":     true,
	"codeBlock":      true,
	"hardBreak":      true,
	"horizontalRule": true,
	"image":          true,
	"text":           true,
}

var allowedMarkTypes = map[string]bool{
	"bold":   true,
	"italic": true,
	"strike": true,
	"code":   true,
	"link":   true,
}

// SanitizeDocument validates and copies a Tiptap document. The returned map
// contains only the nodes and attributes that the public renderer supports.
func SanitizeDocument(input map[string]any) (map[string]any, error) {
	s := documentSanitizer{}
	node, err := s.sanitizeNode(input, 0)
	if err != nil {
		return nil, err
	}
	return node, nil
}

type documentSanitizer struct{ nodes int }

func (s *documentSanitizer) sanitizeNode(input map[string]any, depth int) (map[string]any, error) {
	if depth > maxDocumentDepth {
		return nil, fmt.Errorf("content document is too deeply nested")
	}
	s.nodes++
	if s.nodes > maxDocumentNodes {
		return nil, fmt.Errorf("content document contains too many nodes")
	}

	nodeType, ok := input["type"].(string)
	if !ok || !allowedNodeTypes[nodeType] {
		return nil, fmt.Errorf("content node type %q is not allowed", nodeType)
	}
	output := map[string]any{"type": nodeType}
	attrs, err := sanitizeNodeAttrs(nodeType, input["attrs"])
	if err != nil {
		return nil, err
	}
	if attrs != nil {
		output["attrs"] = attrs
	}

	switch nodeType {
	case "text":
		text, ok := input["text"].(string)
		if !ok || text == "" || len(text) > maxTextLength {
			return nil, fmt.Errorf("text node is invalid")
		}
		output["text"] = text
		marks, err := sanitizeMarks(input["marks"])
		if err != nil {
			return nil, err
		}
		if marks != nil {
			output["marks"] = marks
		}
		return output, rejectUnexpectedKeys(input, "type", "text", "marks")
	case "image", "hardBreak", "horizontalRule":
		if content, exists := input["content"]; exists && content != nil {
			return nil, fmt.Errorf("%s cannot contain content", nodeType)
		}
		return output, rejectUnexpectedKeys(input, "type", "attrs")
	default:
		children, err := sanitizeChildren(s, input["content"], depth+1)
		if err != nil {
			return nil, err
		}
		output["content"] = children
		return output, rejectUnexpectedKeys(input, "type", "attrs", "content")
	}
}

func sanitizeChildren(s *documentSanitizer, raw any, depth int) ([]any, error) {
	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("content node must contain an array")
	}
	children := make([]any, 0, len(items))
	for _, item := range items {
		child, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("content child must be an object")
		}
		clean, err := s.sanitizeNode(child, depth)
		if err != nil {
			return nil, err
		}
		children = append(children, clean)
	}
	return children, nil
}

func sanitizeNodeAttrs(nodeType string, raw any) (map[string]any, error) {
	if raw == nil {
		return nil, nil
	}
	attrs, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s attributes are invalid", nodeType)
	}
	output := make(map[string]any, len(attrs))
	allowed := map[string]bool{}
	switch nodeType {
	case "heading":
		allowed["level"] = true
		level, ok := numberAsInt(attrs["level"])
		if !ok || level < 1 || level > 6 {
			return nil, fmt.Errorf("heading level is invalid")
		}
		output["level"] = attrs["level"]
	case "orderedList":
		allowed["order"] = true
		if value, exists := attrs["order"]; exists {
			order, ok := numberAsInt(value)
			if !ok || order < 1 || order > 1000000 {
				return nil, fmt.Errorf("ordered list start is invalid")
			}
			output["order"] = value
		}
	case "codeBlock":
		allowed["language"] = true
		if language, exists := attrs["language"]; exists {
			value, ok := language.(string)
			if !ok || len(value) > 64 {
				return nil, fmt.Errorf("code language is invalid")
			}
			output["language"] = value
		}
	case "image":
		allowed["src"], allowed["alt"], allowed["title"] = true, true, true
		src, ok := attrs["src"].(string)
		if !ok || ValidateURL(src, false) != nil {
			return nil, fmt.Errorf("image source is invalid")
		}
		output["src"] = src
		for _, key := range []string{"alt", "title"} {
			if value, exists := attrs[key]; exists {
				text, ok := value.(string)
				if !ok || len(text) > 512 {
					return nil, fmt.Errorf("image %s is invalid", key)
				}
				output[key] = text
			}
		}
	default:
		return nil, fmt.Errorf("%s does not accept attributes", nodeType)
	}
	for key := range attrs {
		if !allowed[key] {
			return nil, fmt.Errorf("%s attribute %q is not allowed", nodeType, key)
		}
	}
	return output, nil
}

func sanitizeMarks(raw any) ([]any, error) {
	if raw == nil {
		return nil, nil
	}
	marks, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("text marks are invalid")
	}
	output := make([]any, 0, len(marks))
	for _, rawMark := range marks {
		mark, ok := rawMark.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("text mark must be an object")
		}
		markType, ok := mark["type"].(string)
		if !ok || !allowedMarkTypes[markType] {
			return nil, fmt.Errorf("text mark type is not allowed")
		}
		clean := map[string]any{"type": markType}
		if markType == "link" {
			attrs, ok := mark["attrs"].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("link attributes are required")
			}
			href, ok := attrs["href"].(string)
			if !ok || ValidateURL(href, true) != nil {
				return nil, fmt.Errorf("link URL is invalid")
			}
			linkAttrs := map[string]any{"href": href}
			for _, key := range []string{"target", "rel"} {
				if value, exists := attrs[key]; exists {
					text, ok := value.(string)
					if !ok || len(text) > 256 {
						return nil, fmt.Errorf("link %s is invalid", key)
					}
					linkAttrs[key] = text
				}
			}
			for key := range attrs {
				if key != "href" && key != "target" && key != "rel" {
					return nil, fmt.Errorf("link attribute %q is not allowed", key)
				}
			}
			clean["attrs"] = linkAttrs
		} else if attrs, exists := mark["attrs"]; exists && attrs != nil {
			return nil, fmt.Errorf("%s does not accept attributes", markType)
		}
		if err := rejectUnexpectedKeys(mark, "type", "attrs"); err != nil {
			return nil, err
		}
		output = append(output, clean)
	}
	return output, nil
}

func rejectUnexpectedKeys(input map[string]any, allowed ...string) error {
	keys := make(map[string]bool, len(allowed))
	for _, key := range allowed {
		keys[key] = true
	}
	for key := range input {
		if !keys[key] {
			return fmt.Errorf("content field %q is not allowed", key)
		}
	}
	return nil
}

func numberAsInt(value any) (int, bool) {
	switch number := value.(type) {
	case int:
		return number, true
	case int64:
		return int(number), true
	case float64:
		return int(number), number == float64(int(number))
	default:
		return 0, false
	}
}

// ValidateURL rejects dangerous schemes and targets that should never be
// linked from public site content. The application never fetches this URL.
func ValidateURL(raw string, allowMailto bool) error {
	if len(raw) == 0 || len(raw) > 2048 || strings.TrimSpace(raw) != raw {
		return fmt.Errorf("URL is empty or too long")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.User != nil {
		return fmt.Errorf("URL is invalid")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme == "mailto" {
		if !allowMailto || parsed.Opaque == "" || strings.ContainsAny(parsed.Opaque, "\r\n") {
			return fmt.Errorf("mailto URL is not allowed")
		}
		return nil
	}
	if scheme != "http" && scheme != "https" || parsed.Hostname() == "" || parsed.Host == "" {
		return fmt.Errorf("URL scheme is not allowed")
	}
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".internal") || strings.HasSuffix(host, ".local") {
		return fmt.Errorf("private host is not allowed")
	}
	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsMulticast()) {
		return fmt.Errorf("private host is not allowed")
	}
	return nil
}
