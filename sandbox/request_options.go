package sandbox

import (
	"io"
	"net/http"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

func transportOptions(options structs.RequestOptions) transport.RequestOptions {
	headers := make(http.Header, len(options.Headers))
	for name, value := range options.Headers {
		headers.Set(name, value)
	}
	result := transport.RequestOptions{
		Headers:      headers,
		Timeout:      options.Timeout,
		DisableRetry: options.DisableRetry,
	}
	if options.Retry != nil {
		result.Retry = &transport.RetryConfig{
			MaxRetries: options.Retry.MaxRetries,
			BaseDelay:  options.Retry.BaseDelay,
			MaxDelay:   options.Retry.MaxDelay,
		}
	}
	return result
}

func responseStatusError(response *http.Response, method, endpoint string) error {
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	return responseAPIError(response, method, endpoint, body)
}
