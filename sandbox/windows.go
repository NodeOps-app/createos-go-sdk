package sandbox

import (
	"context"
	"net/http"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// WindowsService provides desktop-window operations.
type WindowsService struct{ computer *ComputerService }

// List returns visible windows.
func (s *WindowsService) List(ctx context.Context, options structs.ComputerListWindowsOptions) ([]structs.ComputerWindow, error) {
	values := screenQuery(options.ScreenID)
	if options.Application != "" {
		values.Set("application", options.Application)
	}
	httpOptions := transportOptions(options.RequestOptions)
	httpOptions.Query = values
	var windows []structs.ComputerWindow
	err := s.computer.instance.transport.Do(ctx, http.MethodGet, s.computer.path("/windows"), httpOptions, &windows)
	return windows, err
}

// Current returns the active window.
func (s *WindowsService) Current(ctx context.Context, options structs.ComputerScreenOptions) (structs.ComputerWindow, error) {
	return s.getAt(ctx, "/windows/current", options)
}

// Get returns one window.
func (s *WindowsService) Get(ctx context.Context, windowID string, options structs.ComputerScreenOptions) (structs.ComputerWindow, error) {
	return s.getAt(ctx, "/windows/"+transport.EncodePath(windowID), options)
}

func (s *WindowsService) getAt(ctx context.Context, suffix string, options structs.ComputerScreenOptions) (structs.ComputerWindow, error) {
	var window structs.ComputerWindow
	err := s.computer.do(ctx, http.MethodGet, suffix, options, nil, &window)
	return window, err
}

// Geometry returns a window's position and dimensions.
func (s *WindowsService) Geometry(ctx context.Context, windowID string, options structs.ComputerScreenOptions) (structs.ComputerWindowGeometry, error) {
	var geometry structs.ComputerWindowGeometry
	err := s.computer.do(ctx, http.MethodGet, "/windows/"+transport.EncodePath(windowID)+"/geometry", options, nil, &geometry)
	return geometry, err
}

// Focus focuses a window.
func (s *WindowsService) Focus(ctx context.Context, windowID string, options structs.ComputerScreenOptions) error {
	return s.action(ctx, windowID, "focus", options, nil)
}

// Move moves a window.
func (s *WindowsService) Move(ctx context.Context, windowID string, request structs.ComputerWindowMoveRequest, options structs.ComputerScreenOptions) error {
	return s.action(ctx, windowID, "move", options, request)
}

// Resize resizes a window.
func (s *WindowsService) Resize(ctx context.Context, windowID string, request structs.ComputerWindowResizeRequest, options structs.ComputerScreenOptions) error {
	return s.action(ctx, windowID, "resize", options, request)
}

// Maximize maximizes a window.
func (s *WindowsService) Maximize(ctx context.Context, windowID string, options structs.ComputerScreenOptions) error {
	return s.action(ctx, windowID, "maximize", options, nil)
}

// Minimize minimizes a window.
func (s *WindowsService) Minimize(ctx context.Context, windowID string, options structs.ComputerScreenOptions) error {
	return s.action(ctx, windowID, "minimize", options, nil)
}

// Restore restores a window.
func (s *WindowsService) Restore(ctx context.Context, windowID string, options structs.ComputerScreenOptions) error {
	return s.action(ctx, windowID, "restore", options, nil)
}

// Close closes a window.
func (s *WindowsService) Close(ctx context.Context, windowID string, options structs.ComputerScreenOptions) error {
	var response structs.OKResponse
	httpOptions := transportOptions(options.RequestOptions)
	httpOptions.Query = screenQuery(options.ScreenID)
	return s.computer.instance.transport.Do(ctx, http.MethodDelete, s.computer.path("/windows/"+transport.EncodePath(windowID)), httpOptions, &response)
}

func (s *WindowsService) action(ctx context.Context, windowID, action string, options structs.ComputerScreenOptions, body any) error {
	return s.computer.doOK(ctx, http.MethodPost, "/windows/"+transport.EncodePath(windowID)+"/"+action, options, body)
}
