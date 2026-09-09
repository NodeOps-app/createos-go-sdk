// Package redact removes credentials from observability metadata.
package redact

import (
	"net/http"
	"net/url"
	"strings"
)

const replacement = "<redacted>"

var sensitiveHeaders = map[string]struct{}{
	"authorization":       {},
	"proxy-authorization": {},
	"cookie":              {},
	"set-cookie":          {},
	"x-api-key":           {},
	"x-auth-token":        {},
	"x-csrf-token":        {},
}

var sensitiveQueryParts = []string{"token", "secret", "password", "api_key", "apikey", "credential", "signature"}

// Headers returns a cloned header map with credential values removed.
func Headers(headers http.Header) http.Header {
	result := headers.Clone()
	for name := range result {
		if IsSensitiveHeader(name) {
			result[name] = []string{replacement}
		}
	}
	return result
}

// IsSensitiveHeader reports whether a header commonly carries credentials.
func IsSensitiveHeader(name string) bool {
	_, ok := sensitiveHeaders[strings.ToLower(name)]
	return ok
}

// URL redacts credentials embedded in a URL or sensitive query parameters.
func URL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if u.User != nil {
		u.User = url.User(replacement)
	}
	query := u.Query()
	for key := range query {
		lower := strings.ToLower(key)
		for _, part := range sensitiveQueryParts {
			if strings.Contains(lower, part) {
				query.Set(key, replacement)
				break
			}
		}
	}
	u.RawQuery = query.Encode()
	return u.String()
}
