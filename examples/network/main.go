// Command network connects a sandbox to a private overlay network and verifies
// that the sandbox appears in the network membership view.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/NodeOps-app/createos-go-sdk/sandbox"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	client, err := sandbox.NewClient()
	if err != nil {
		return err
	}

	network, err := client.Networks().Create(ctx, structs.NetworkCreateRequest{
		Name: fmt.Sprintf("go-sdk-%d", time.Now().Unix()),
	})
	if err != nil {
		return fmt.Errorf("create network: %w", err)
	}
	fmt.Printf("created network: %s\n", network.ID)
	defer deleteNetwork(ctx, client, network.ID)

	instance, err := client.CreateSandbox(ctx, structs.CreateSandboxRequest{
		Shape:  "s-1vcpu-1gb",
		RootFS: "devbox:1",
	})
	if err != nil {
		return fmt.Errorf("create sandbox: %w", err)
	}
	fmt.Printf("created sandbox: %s\n", instance.ID())
	defer destroy(ctx, instance)

	if err := instance.AttachNetwork(ctx, network.ID); err != nil {
		return fmt.Errorf("attach network: %w", err)
	}
	defer detachNetwork(ctx, instance, network.ID)

	connected, err := client.Networks().Get(ctx, network.ID)
	if err != nil {
		return fmt.Errorf("get network: %w", err)
	}
	for _, member := range connected.Members {
		if member.SandboxID == instance.ID() {
			fmt.Printf("verified member: sandbox=%s ip=%s status=%s\n",
				member.SandboxID, member.IPAddress, member.Status)
			return nil
		}
	}
	return fmt.Errorf("sandbox %s not found in network %s", instance.ID(), network.ID)
}

func detachNetwork(parent context.Context, instance *sandbox.Instance, networkID string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), 30*time.Second)
	defer cancel()
	if err := instance.DetachNetwork(ctx, networkID); err != nil {
		log.Printf("cleanup: detach network %s: %v", networkID, err)
		return
	}
	fmt.Printf("detached network: %s\n", networkID)
}

func destroy(parent context.Context, instance *sandbox.Instance) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), 30*time.Second)
	defer cancel()
	if err := instance.Destroy(ctx); err != nil {
		log.Printf("cleanup: destroy sandbox %s: %v", instance.ID(), err)
		return
	}
	fmt.Printf("destroyed sandbox: %s\n", instance.ID())
}

func deleteNetwork(parent context.Context, client *sandbox.Client, networkID string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), 30*time.Second)
	defer cancel()
	if err := client.Networks().Delete(ctx, networkID); err != nil {
		log.Printf("cleanup: delete network %s: %v", networkID, err)
		return
	}
	fmt.Printf("deleted network: %s\n", networkID)
}
