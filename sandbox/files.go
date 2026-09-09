package sandbox

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
)

// FilesService provides file transfer for one sandbox.
type FilesService struct{ instance *Instance }

// Upload writes data to an absolute sandbox path.
func (s *FilesService) Upload(ctx context.Context, path string, data io.Reader) error {
	response, err := s.instance.transport.DoRaw(ctx, http.MethodPut, s.instance.path("/files"), transport.RequestOptions{
		Query: url.Values{"path": {path}}, RawBody: data, ContentType: "application/octet-stream",
	})
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
func (s *FilesService) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	response, err := s.instance.transport.DoRaw(ctx, http.MethodGet, s.instance.path("/files"), transport.RequestOptions{
		Query: url.Values{"path": {path}},
	})
	if err != nil {
		return nil, fmt.Errorf("download %q from sandbox %q: %w", path, s.instance.ID(), err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, responseStatusError(response, http.MethodGet, s.instance.path("/files"))
	}
	return response.Body, nil
}
