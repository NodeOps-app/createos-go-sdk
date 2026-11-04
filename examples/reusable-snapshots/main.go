// Command reusable-snapshots captures a disk snapshot and restores a sandbox from it.
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
	ctx := context.Background()
	client, err := sandbox.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	source, err := client.CreateSandbox(ctx, structs.CreateSandboxRequest{
		Shape: "s-1vcpu-1gb", RootFS: "alpine:3.20",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer destroy(source)

	snapshot, err := source.CreateSnapshot(ctx, structs.CreateSnapshotRequest{
		Type: structs.SnapshotTypeDisk, Name: "configured-workspace",
	})
	if err != nil {
		log.Printf("create snapshot: %v", err)
		return
	}
	defer func() {
		latest, err := client.Snapshots().Get(context.Background(), snapshot.ID)
		if err != nil {
			log.Printf("get snapshot for cleanup: %v", err)
			return
		}
		if latest.Status == structs.SnapshotStatusReady || latest.Status == structs.SnapshotStatusFailed {
			if _, err := client.Snapshots().Delete(context.Background(), snapshot.ID); err != nil {
				log.Printf("delete snapshot: %v", err)
			}
		}
	}()
	snapshot, err = waitUntilReady(ctx, client, snapshot.ID, 5*time.Minute)
	if err != nil {
		log.Printf("wait for snapshot: %v", err)
		return
	}

	restored, err := client.CreateSandbox(ctx, structs.CreateSandboxRequest{SnapshotID: snapshot.ID})
	if err != nil {
		log.Printf("restore sandbox: %v", err)
		return
	}
	defer destroy(restored)
	fmt.Printf("restored sandbox %s from snapshot %s\n", restored.ID(), snapshot.ID)
}

func waitUntilReady(ctx context.Context, client *sandbox.Client, id string, timeout time.Duration) (*structs.SnapshotView, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snapshot, err := client.Snapshots().Get(ctx, id)
		if err != nil {
			return nil, err
		}
		switch snapshot.Status {
		case structs.SnapshotStatusReady:
			return snapshot, nil
		case structs.SnapshotStatusFailed:
			return nil, fmt.Errorf("snapshot %s failed", id)
		case structs.SnapshotStatusCreating, structs.SnapshotStatusDeleting:
			// Continue polling until the snapshot reaches a terminal state.
		}
		time.Sleep(time.Second)
	}
	return nil, fmt.Errorf("timed out waiting for snapshot %s", id)
}

func destroy(instance *sandbox.Instance) {
	if err := instance.Destroy(context.Background()); err != nil {
		log.Printf("destroy sandbox %s: %v", instance.ID(), err)
	}
}
