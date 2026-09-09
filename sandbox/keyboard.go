package sandbox

import (
	"context"
	"net/http"

	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// KeyboardService provides keyboard operations.
type KeyboardService struct{ computer *ComputerService }

// Type types text into the active application.
func (s *KeyboardService) Type(ctx context.Context, request structs.ComputerTypeRequest, options structs.ComputerScreenOptions) error {
	return s.computer.doOK(ctx, http.MethodPost, "/keyboard/type", options, request)
}

// Press presses and releases a key combination.
func (s *KeyboardService) Press(ctx context.Context, keys []string, options structs.ComputerScreenOptions) error {
	return s.computer.doOK(ctx, http.MethodPost, "/keyboard/press", options, structs.ComputerPressRequest{Keys: keys})
}

// Down presses keys without releasing them.
func (s *KeyboardService) Down(ctx context.Context, keys []string, options structs.ComputerScreenOptions) error {
	return s.computer.doOK(ctx, http.MethodPost, "/keyboard/down", options, structs.ComputerPressRequest{Keys: keys})
}

// Up releases keys.
func (s *KeyboardService) Up(ctx context.Context, keys []string, options structs.ComputerScreenOptions) error {
	return s.computer.doOK(ctx, http.MethodPost, "/keyboard/up", options, structs.ComputerPressRequest{Keys: keys})
}
