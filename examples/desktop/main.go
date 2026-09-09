// Command desktop demonstrates screenshots, basic computer controls, and a
// temporary noVNC connection to a graphical sandbox.
package main

import (
	"context"
	"errors"
	"fmt"
	"image/png"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/NodeOps-app/createos-go-sdk/sandbox"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

const primaryScreen = structs.ComputerScreen0

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx)
	stop()
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) (runErr error) {
	client, err := sandbox.NewClient()
	if err != nil {
		return err
	}
	instance, err := client.CreateSandbox(ctx, structs.CreateSandboxRequest{
		Shape:          "s-2vcpu-4gb",
		RootFS:         "desktop:1",
		IngressEnabled: true,
	})
	if err != nil {
		return err
	}
	fmt.Println("created:", instance.ID())
	defer func(cleanupParent context.Context) {
		cleanupCtx, cancel := context.WithTimeout(cleanupParent, 30*time.Second)
		defer cancel()
		if err := instance.Destroy(cleanupCtx); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("destroy sandbox: %w", err))
			return
		}
		fmt.Println("destroyed")
	}(context.WithoutCancel(ctx))

	computer := instance.Computer()
	screenOptions := structs.ComputerScreenOptions{ScreenID: primaryScreen}
	fmt.Println("\n[1/5] reading primary screen...")
	geometry, err := waitForDesktop(ctx, computer, screenOptions)
	if err != nil {
		return err
	}
	screens, err := computer.Screens().List(ctx)
	if err != nil {
		return err
	}
	primary, err := computer.Screens().Get(ctx, primaryScreen)
	if err != nil {
		return err
	}
	fmt.Printf("      geometry: %dx%d\n", geometry.Width, geometry.Height)
	for index, screen := range screens {
		if index == 0 {
			fmt.Print("      screens: ")
		} else {
			fmt.Print(", ")
		}
		fmt.Print(screen.ScreenID)
	}
	fmt.Println()
	fmt.Printf("      primary display: %s, noVNC port %d\n", primary.Display, primary.NoVNCPort)

	fmt.Println("\n[2/5] capturing PNG screenshots...")
	fullWidth, fullHeight, err := screenshotSize(ctx, computer, structs.ComputerScreenshotOptions{
		ComputerScreenOptions: structs.ComputerScreenOptions{
			ScreenID:       primaryScreen,
			RequestOptions: structs.RequestOptions{Timeout: 45 * time.Second},
		},
	})
	if err != nil {
		return err
	}
	fmt.Printf("      full screenshot: %dx%d\n", fullWidth, fullHeight)

	regionWidth, regionHeight := 240, 160
	zero := 0
	cropWidth, cropHeight, err := screenshotSize(ctx, computer, structs.ComputerScreenshotOptions{
		ComputerScreenOptions: structs.ComputerScreenOptions{
			ScreenID:       primaryScreen,
			RequestOptions: structs.RequestOptions{Timeout: 45 * time.Second},
		},
		X: &zero, Y: &zero, Width: regionWidth, Height: regionHeight,
	})
	if err != nil {
		return err
	}
	fmt.Printf("      region screenshot: %dx%d\n", cropWidth, cropHeight)

	fmt.Println("\n[3/5] moving cursor and round-tripping clipboard...")
	target := structs.ComputerPoint{
		X: min(max(geometry.Width/3, 10), geometry.Width-1),
		Y: min(max(geometry.Height/3, 10), geometry.Height-1),
	}
	if err := computer.Mouse().Move(ctx, target, screenOptions); err != nil {
		return err
	}
	cursor, err := computer.Cursor(ctx, screenOptions)
	if err != nil {
		return err
	}
	fmt.Printf("      cursor: %d,%d\n", cursor.X, cursor.Y)

	clipboardText := "CreateOS desktop " + instance.ID()
	if err := computer.SetClipboard(ctx, clipboardText, screenOptions); err != nil {
		return err
	}
	clipboard, err := computer.Clipboard(ctx, screenOptions)
	if err != nil {
		return err
	}
	fmt.Println("      clipboard:", clipboard.Text)

	fmt.Println("\n[4/5] opening a URL in the desktop browser...")
	const targetURL = "https://example.com"
	if err := computer.Open(ctx, structs.ComputerOpenRequest{Target: targetURL}, screenOptions); err != nil {
		return err
	}
	fmt.Println("      opened:", targetURL)

	fmt.Println("\n[5/5] creating live noVNC connection...")
	connection, err := computer.Screens().Connect(ctx, primaryScreen)
	if err != nil {
		return err
	}
	fmt.Println("      screen:", connection.ScreenID)
	fmt.Println("      expires:", connection.ExpiresAt)
	if connection.URL == "" {
		fmt.Println("      noVNC URL: (not available)")
	} else {
		fmt.Println("      noVNC URL:", connection.URL)
	}

	if len(screens) == 0 || primary.ScreenID != primaryScreen {
		return errors.New("primary screen was not listed")
	}
	if fullWidth != geometry.Width || fullHeight != geometry.Height {
		return errors.New("full screenshot dimensions did not match screen geometry")
	}
	if cropWidth != regionWidth || cropHeight != regionHeight {
		return errors.New("region screenshot dimensions did not match request")
	}
	if cursor != target {
		return errors.New("cursor did not move to requested coordinates")
	}
	if clipboard.Text != clipboardText {
		return errors.New("clipboard round trip failed")
	}
	if !strings.HasPrefix(connection.URL, "https://") {
		return errors.New("noVNC connection did not include a public HTTPS URL")
	}
	fmt.Println("\nverified end-to-end: desktop computer API and noVNC connection")
	return nil
}

func waitForDesktop(ctx context.Context, computer *sandbox.ComputerService, options structs.ComputerScreenOptions) (structs.ComputerScreenGeometry, error) {
	waitCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	var lastErr error
	for {
		geometry, err := computer.Screen(waitCtx, options)
		if err == nil {
			return geometry, nil
		}
		lastErr = err
		timer := time.NewTimer(2 * time.Second)
		select {
		case <-waitCtx.Done():
			timer.Stop()
			return structs.ComputerScreenGeometry{}, fmt.Errorf("wait for desktop: %w", errors.Join(waitCtx.Err(), lastErr))
		case <-timer.C:
		}
	}
}

func screenshotSize(ctx context.Context, computer *sandbox.ComputerService, options structs.ComputerScreenshotOptions) (width, height int, resultErr error) {
	body, err := computer.Screenshot(ctx, options)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err := body.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close screenshot response: %w", err))
		}
	}()
	config, err := png.DecodeConfig(body)
	if err != nil {
		return 0, 0, fmt.Errorf("decode screenshot PNG: %w", err)
	}
	return config.Width, config.Height, nil
}
