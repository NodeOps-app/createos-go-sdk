package sandbox

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

// ClientOption configures a Client.
type ClientOption func(*clientConfig) error

// WithAPIKey sets the API key used for authenticated requests.
func WithAPIKey(apiKey string) ClientOption {
	return func(configuration *clientConfig) error {
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			return errors.New("api key must not be empty")
		}
		configuration.apiKey = apiKey
		return nil
	}
}

// WithBaseURL overrides the CreateOS control-plane URL.
func WithBaseURL(baseURL string) ClientOption {
	return func(configuration *clientConfig) error {
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			return errors.New("base URL must not be empty")
		}
		configuration.baseURL = baseURL
		return nil
	}
}

// WithTimeout sets the default timeout for HTTP requests.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(configuration *clientConfig) error {
		if timeout <= 0 {
			return errors.New("timeout must be positive")
		}
		configuration.timeout = timeout
		return nil
	}
}

// WithHTTPClient uses client for HTTP requests. The transport shallow-clones
// the client before installing its redirect policy. Any Timeout already set on
// the supplied client remains an independent upper bound.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(configuration *clientConfig) error {
		if client == nil {
			return errors.New("HTTP client must not be nil")
		}
		configuration.httpClient = client
		return nil
	}
}

// WithUserAgent overrides the value sent in the User-Agent header.
func WithUserAgent(userAgent string) ClientOption {
	return func(configuration *clientConfig) error {
		userAgent = strings.TrimSpace(userAgent)
		if userAgent == "" {
			return errors.New("user agent must not be empty")
		}
		configuration.userAgent = userAgent
		return nil
	}
}

// WithRetry configures retries for retryable requests.
func WithRetry(maxRetries int, baseDelay, maxDelay time.Duration) ClientOption {
	return func(configuration *clientConfig) error {
		if maxRetries < 0 {
			return errors.New("maximum retries must not be negative")
		}
		if baseDelay <= 0 {
			return errors.New("base retry delay must be positive")
		}
		if maxDelay < baseDelay {
			return errors.New("maximum retry delay must not be less than base retry delay")
		}
		configuration.retry = retryConfig{
			maxRetries: maxRetries,
			baseDelay:  baseDelay,
			maxDelay:   maxDelay,
		}
		return nil
	}
}

// WithoutRetry disables automatic retries.
func WithoutRetry() ClientOption {
	return func(configuration *clientConfig) error {
		configuration.retry.maxRetries = 0
		return nil
	}
}
