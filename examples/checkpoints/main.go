// Command checkpoints captures and restores local directory history.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/NodeOps-app/createos-go-sdk/sandbox"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

const checkpointPath = "/workspace/checkpoint-demo"

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
	instance, err := client.CreateSandbox(ctx, structs.CreateSandboxRequest{
		Shape: "s-1vcpu-1gb", RootFS: "devbox:1",
	})
	if err != nil {
		return err
	}
	defer destroy(ctx, instance)

	if _, err = instance.Shell(ctx, "mkdir -p "+checkpointPath+" && printf 'version one' > "+checkpointPath+"/value", structs.ExecOptions{}); err != nil {
		return err
	}
	checkpoint, err := instance.Checkpoints().Create(ctx, structs.CreateCheckpointRequest{
		Path: checkpointPath, Name: "before-update",
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := instance.Checkpoints().Delete(ctx, checkpoint.ID); err != nil {
			log.Printf("delete checkpoint: %v", err)
		}
	}()

	if _, err = instance.Shell(ctx, "printf 'version two' > "+checkpointPath+"/value", structs.ExecOptions{}); err != nil {
		return err
	}
	if _, err = instance.Checkpoints().Restore(ctx, checkpoint.ID); err != nil {
		return err
	}
	result, err := instance.Shell(ctx, "cat "+checkpointPath+"/value", structs.ExecOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("restored %s: %s\n", checkpoint.ID, result.Result.StandardOutput)
	return nil
}

func destroy(ctx context.Context, instance *sandbox.Instance) {
	if err := instance.Destroy(ctx); err != nil {
		log.Printf("destroy sandbox %s: %v", instance.ID(), err)
	}
}
