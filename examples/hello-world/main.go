// Command hello-world creates a sandbox, runs a command, and destroys the
// sandbox before exiting.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/NodeOps-app/createos-go-sdk/sandbox"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

func main() {
	ctx := context.Background()
	client, err := sandbox.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	instance, err := client.CreateSandbox(ctx, structs.CreateSandboxRequest{
		Shape:  "s-4vcpu-4gb",
		RootFS: "devbox:1",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := instance.Destroy(context.Background()); err != nil {
			log.Printf("destroy sandbox: %v", err)
		}
	}()

	response, err := instance.RunCommand(ctx, structs.RunCommandRequest{
		Command:   "sh",
		Arguments: []string{"-c", `printf "Go says hello from $(uname -m)\n"`},
	}, structs.ExecOptions{})
	if err != nil {
		log.Printf("run command: %v", err)
		return
	}
	fmt.Print(response.Result.StandardOutput)
}
