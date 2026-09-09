package sandbox

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// ComputerService provides desktop computer-use operations.
type ComputerService struct {
	instance *Instance
	mouse    *MouseService
	keyboard *KeyboardService
	windows  *WindowsService
	screens  *ScreensService
}

func newComputerService(instance *Instance) *ComputerService {
	service := &ComputerService{instance: instance}
	service.mouse = &MouseService{computer: service}
	service.keyboard = &KeyboardService{computer: service}
	service.windows = &WindowsService{computer: service}
	service.screens = &ScreensService{computer: service}
	return service
}

func (s *ComputerService) path(suffix string) string { return s.instance.path("/computer" + suffix) }

// Mouse returns mouse operations.
func (s *ComputerService) Mouse() *MouseService { return s.mouse }

// Keyboard returns keyboard operations.
func (s *ComputerService) Keyboard() *KeyboardService { return s.keyboard }

// Windows returns desktop-window operations.
func (s *ComputerService) Windows() *WindowsService { return s.windows }

// Screens returns screen and connection operations.
func (s *ComputerService) Screens() *ScreensService { return s.screens }

// Screenshot captures PNG bytes. The caller must close the returned body.
func (s *ComputerService) Screenshot(ctx context.Context, options structs.ComputerScreenshotOptions) (io.ReadCloser, error) {
	values := screenQuery(options.ScreenID)
	if options.WindowID != "" {
		values.Set("window_id", options.WindowID)
	}
	setOptionalInt(values, "x", options.X)
	setOptionalInt(values, "y", options.Y)
	if options.Width != 0 {
		values.Set("width", strconv.Itoa(options.Width))
	}
	if options.Height != 0 {
		values.Set("height", strconv.Itoa(options.Height))
	}
	httpOptions := transportOptions(options.RequestOptions)
	httpOptions.Query = values
	response, err := s.instance.transport.DoRaw(ctx, http.MethodGet, s.path("/screenshot"), httpOptions)
	if err != nil {
		return nil, fmt.Errorf("capture sandbox screenshot: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, responseStatusError(response, http.MethodGet, s.path("/screenshot"))
	}
	return response.Body, nil
}

// Screen returns active screen dimensions.
func (s *ComputerService) Screen(ctx context.Context, options structs.ComputerScreenOptions) (structs.ComputerScreenGeometry, error) {
	var result structs.ComputerScreenGeometry
	err := s.do(ctx, http.MethodGet, "/screen", options, nil, &result)
	return result, err
}

// Cursor returns current cursor coordinates.
func (s *ComputerService) Cursor(ctx context.Context, options structs.ComputerScreenOptions) (structs.ComputerPoint, error) {
	var result structs.ComputerPoint
	err := s.do(ctx, http.MethodGet, "/cursor", options, nil, &result)
	return result, err
}

// Clipboard returns current clipboard text.
func (s *ComputerService) Clipboard(ctx context.Context, options structs.ComputerScreenOptions) (structs.ComputerClipboard, error) {
	var result structs.ComputerClipboard
	err := s.do(ctx, http.MethodGet, "/clipboard", options, nil, &result)
	return result, err
}

// SetClipboard replaces clipboard text.
func (s *ComputerService) SetClipboard(ctx context.Context, text string, options structs.ComputerScreenOptions) error {
	return s.doOK(ctx, http.MethodPut, "/clipboard", options, structs.ComputerClipboard{Text: text})
}

// Open opens a URL or desktop target.
func (s *ComputerService) Open(ctx context.Context, request structs.ComputerOpenRequest, options structs.ComputerScreenOptions) error {
	return s.doOK(ctx, http.MethodPost, "/open", options, request)
}

// Launch starts an installed desktop application.
func (s *ComputerService) Launch(ctx context.Context, request structs.ComputerLaunchRequest, options structs.ComputerScreenOptions) error {
	return s.doOK(ctx, http.MethodPost, "/launch", options, request)
}

func (s *ComputerService) do(ctx context.Context, method, suffix string, options structs.ComputerScreenOptions, body, out any) error {
	httpOptions := transportOptions(options.RequestOptions)
	httpOptions.Query = screenQuery(options.ScreenID)
	httpOptions.Body = body
	if err := s.instance.transport.Do(ctx, method, s.path(suffix), httpOptions, out); err != nil {
		return fmt.Errorf("computer operation %s: %w", suffix, err)
	}
	return nil
}

func (s *ComputerService) doOK(ctx context.Context, method, suffix string, options structs.ComputerScreenOptions, body any) error {
	var response structs.OKResponse
	return s.do(ctx, method, suffix, options, body, &response)
}

func screenQuery(screenID structs.ComputerScreenID) url.Values {
	values := make(url.Values)
	if screenID != "" {
		values.Set("screen_id", string(screenID))
	}
	return values
}

func setOptionalInt(values url.Values, key string, value *int) {
	if value != nil {
		values.Set(key, strconv.Itoa(*value))
	}
}
