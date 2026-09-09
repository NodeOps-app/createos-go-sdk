package sandbox

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultBaseURL   = "https://api.sb.createos.sh"
	defaultTimeout   = 60 * time.Second
	defaultUserAgent = "createos-go-sdk/0.0.1"
)

type clientConfig struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	userAgent  string
	timeout    time.Duration
	retry      retryConfig
}

type retryConfig struct {
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

func resolveConfig(options ...ClientOption) (clientConfig, error) {
	configuration := clientConfig{
		apiKey:    strings.TrimSpace(os.Getenv("CREATEOS_SANDBOX_API_KEY")),
		baseURL:   strings.TrimSpace(os.Getenv("CREATEOS_SANDBOX_BASE_URL")),
		userAgent: defaultUserAgent,
		timeout:   defaultTimeout,
		retry: retryConfig{
			maxRetries: 2,
			baseDelay:  500 * time.Millisecond,
			maxDelay:   30 * time.Second,
		},
	}
	if configuration.baseURL == "" {
		configuration.baseURL = defaultBaseURL
	}

	for index, option := range options {
		if option == nil {
			return clientConfig{}, fmt.Errorf("client option %d is nil", index)
		}
		if err := option(&configuration); err != nil {
			return clientConfig{}, fmt.Errorf("apply client option %d: %w", index, err)
		}
	}

	parsedURL, err := url.Parse(configuration.baseURL)
	if err != nil {
		return clientConfig{}, fmt.Errorf("parse base URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return clientConfig{}, errors.New("base URL scheme must be http or https")
	}
	if parsedURL.Host == "" {
		return clientConfig{}, errors.New("base URL must include a host")
	}
	if parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return clientConfig{}, errors.New("base URL must not include a query string or fragment")
	}
	configuration.baseURL = strings.TrimRight(parsedURL.String(), "/")

	if configuration.httpClient == nil {
		configuration.httpClient = &http.Client{}
	}

	return configuration, nil
}
