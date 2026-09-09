package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// FilesService provides file transfer for one sandbox.
type FilesService struct{ instance *Instance }

// Upload writes data to an absolute sandbox path. An optional RequestOptions
// value can override the timeout for the complete upload.
func (s *FilesService) Upload(ctx context.Context, path string, data io.Reader, options ...structs.RequestOptions) error {
	requestOptions, err := fileRequestOptions(options)
	if err != nil {
		return fmt.Errorf("upload %q to sandbox %q: %w", path, s.instance.ID(), err)
	}
	requestOptions.Query = url.Values{"path": {path}}
	requestOptions.RawBody = data
	requestOptions.ContentType = "application/octet-stream"
	response, err := s.instance.transport.DoRaw(ctx, http.MethodPut, s.instance.path("/files"), requestOptions)
	if err != nil {
		return fmt.Errorf("upload %q to sandbox %q: %w", path, s.instance.ID(), err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return responseStatusError(response, http.MethodPut, s.instance.path("/files"))
	}
	return nil
}

// Download opens a file for reading. The caller must close the returned body.
// An optional RequestOptions value can override the timeout; that timeout
// remains active until the body reaches EOF or is closed.
func (s *FilesService) Download(ctx context.Context, path string, options ...structs.RequestOptions) (io.ReadCloser, error) {
	requestOptions, err := fileRequestOptions(options)
	if err != nil {
		return nil, fmt.Errorf("download %q from sandbox %q: %w", path, s.instance.ID(), err)
	}
	requestOptions.Query = url.Values{"path": {path}}
	response, err := s.instance.transport.DoRaw(ctx, http.MethodGet, s.instance.path("/files"), requestOptions)
	if err != nil {
		return nil, fmt.Errorf("download %q from sandbox %q: %w", path, s.instance.ID(), err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, responseStatusError(response, http.MethodGet, s.instance.path("/files"))
	}
	return response.Body, nil
}

func fileRequestOptions(options []structs.RequestOptions) (transport.RequestOptions, error) {
	if len(options) > 1 {
		return transport.RequestOptions{}, errors.New("at most one request options value may be provided")
	}
	if len(options) == 0 {
		return transport.RequestOptions{}, nil
	}
	return transportOptions(options[0]), nil
}
