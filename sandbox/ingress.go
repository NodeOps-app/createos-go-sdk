package sandbox

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// PreviewURL returns the public ingress URL for a sandbox port.
func (i *Instance) PreviewURL(port int) (*url.URL, error) {
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}
	data := i.snapshot()
	if !data.IngressEnabled || data.IngressURLTemplate == "" {
		return nil, errors.New("sandbox ingress is not enabled")
	}
	value := strings.ReplaceAll(data.IngressURLTemplate, "<port>", strconv.Itoa(port))
	parsed, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("parse ingress URL: %w", err)
	}
	return parsed, nil
}

// WaitForPort waits until a TCP port is listening inside the sandbox.
func (i *Instance) WaitForPort(ctx context.Context, host string, port int, timeout time.Duration) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}
	if host == "" {
		host = "127.0.0.1"
	}
	if net.ParseIP(host) == nil {
		for _, label := range strings.Split(host, ".") {
			if label == "" || strings.Trim(label, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-") != "" {
				return fmt.Errorf("host %q is not a valid IP address or DNS name", host)
			}
		}
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	seconds := max(1, int((timeout+time.Second-1)/time.Second))
	script := fmt.Sprintf("timeout %d bash -c 'until (echo > /dev/tcp/%s/%d) 2>/dev/null; do sleep 0.25; done'", seconds, host, port)
	response, err := i.RunCommand(ctx, structs.RunCommandRequest{
		Command: "bash", Arguments: []string{"-c", script},
	}, structs.ExecOptions{RequestOptions: structs.RequestOptions{Timeout: timeout + 5*time.Second}})
	if err != nil {
		return err
	}
	if response.Result.ExitCode != 0 {
		return fmt.Errorf("port %d did not become ready within %s: %w", port, timeout, ErrTimeout)
	}
	return nil
}
