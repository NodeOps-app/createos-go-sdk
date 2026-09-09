package sandbox

import (
	"context"
	"net/http"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// ScreensService provides desktop screen operations.
type ScreensService struct{ computer *ComputerService }

// List returns configured screens.
func (s *ScreensService) List(ctx context.Context) ([]structs.ComputerScreen, error) {
	var screens []structs.ComputerScreen
	err := s.computer.instance.transport.Do(ctx, http.MethodGet, s.computer.path("/screens"), transport.RequestOptions{}, &screens)
	return screens, err
}

// Create creates a screen.
func (s *ScreensService) Create(ctx context.Context, request structs.ComputerCreateScreenRequest) (structs.ComputerScreen, error) {
	var screen structs.ComputerScreen
	err := s.computer.instance.transport.Do(ctx, http.MethodPost, s.computer.path("/screens"), transport.RequestOptions{Body: request}, &screen)
	return screen, err
}

// Get returns one screen.
func (s *ScreensService) Get(ctx context.Context, screenID structs.ComputerScreenID) (structs.ComputerScreen, error) {
	var screen structs.ComputerScreen
	err := s.computer.instance.transport.Do(ctx, http.MethodGet, s.screenPath(screenID, ""), transport.RequestOptions{}, &screen)
	return screen, err
}

// Connect returns a temporary noVNC connection.
func (s *ScreensService) Connect(ctx context.Context, screenID structs.ComputerScreenID) (structs.ComputerScreenConnection, error) {
	var connection structs.ComputerScreenConnection
	err := s.computer.instance.transport.Do(ctx, http.MethodGet, s.screenPath(screenID, "/connect"), transport.RequestOptions{}, &connection)
	return connection, err
}

// Resize resizes a screen.
func (s *ScreensService) Resize(ctx context.Context, screenID structs.ComputerScreenID, request structs.ComputerCreateScreenRequest) (structs.ComputerScreen, error) {
	var screen structs.ComputerScreen
	err := s.computer.instance.transport.Do(ctx, http.MethodPost, s.screenPath(screenID, "/resize"), transport.RequestOptions{Body: request}, &screen)
	return screen, err
}

// Delete deletes a screen.
func (s *ScreensService) Delete(ctx context.Context, screenID structs.ComputerScreenID) error {
	var response structs.OKResponse
	return s.computer.instance.transport.Do(ctx, http.MethodDelete, s.screenPath(screenID, ""), transport.RequestOptions{}, &response)
}

func (s *ScreensService) screenPath(screenID structs.ComputerScreenID, suffix string) string {
	return s.computer.path("/screens/" + transport.EncodePath(string(screenID)) + suffix)
}
