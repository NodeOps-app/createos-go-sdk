package sandbox

import (
	"errors"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
)

// ErrTimeout indicates that an SDK operation did not complete within its
// configured wait budget.
var ErrTimeout = errors.New("sandbox operation timed out")

// APIError describes a non-successful response from the CreateOS API.
// Use errors.As to inspect its status code, stable API code, request ID, and
// response headers.
type APIError = transport.ResponseError

// IsAPIError reports whether err contains an API response error.
func IsAPIError(err error) bool {
	var apiError *APIError
	return errors.As(err, &apiError)
}
