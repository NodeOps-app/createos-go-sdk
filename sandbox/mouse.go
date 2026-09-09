package sandbox

import (
	"context"
	"net/http"

	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// MouseService provides mouse operations.
type MouseService struct{ computer *ComputerService }

// Move moves the cursor.
func (s *MouseService) Move(ctx context.Context, point structs.ComputerPoint, options structs.ComputerScreenOptions) error {
	return s.computer.doOK(ctx, http.MethodPost, "/mouse/move", options, point)
}

// Click clicks a mouse button.
func (s *MouseService) Click(ctx context.Context, request structs.ComputerClickRequest, options structs.ComputerScreenOptions) error {
	return s.computer.doOK(ctx, http.MethodPost, "/mouse/click", options, request)
}

// Scroll scrolls the desktop.
func (s *MouseService) Scroll(ctx context.Context, request structs.ComputerScrollRequest, options structs.ComputerScreenOptions) error {
	return s.computer.doOK(ctx, http.MethodPost, "/mouse/scroll", options, request)
}

// Drag drags between two points.
func (s *MouseService) Drag(ctx context.Context, request structs.ComputerDragRequest, options structs.ComputerScreenOptions) error {
	return s.computer.doOK(ctx, http.MethodPost, "/mouse/drag", options, request)
}

// Down presses a mouse button.
func (s *MouseService) Down(ctx context.Context, request structs.ComputerButtonRequest, options structs.ComputerScreenOptions) error {
	return s.computer.doOK(ctx, http.MethodPost, "/mouse/down", options, request)
}

// Up releases a mouse button.
func (s *MouseService) Up(ctx context.Context, request structs.ComputerButtonRequest, options structs.ComputerScreenOptions) error {
	return s.computer.doOK(ctx, http.MethodPost, "/mouse/up", options, request)
}
