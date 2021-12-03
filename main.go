package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

const (
	ActionCreate = "create"
	ActionStart  = "start"
)

func main() {
	dcli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	msgC, errC := dcli.Events(ctx, types.EventsOptions{})

	created := 0
	started := 0

	startedAt := time.Now()
	fmt.Println("Capturing docker events...")

	go func() {
		for {
			select {
			case msg := <-msgC:
				if msg.Type == events.ContainerEventType {
					if msg.Action == ActionCreate {
						created++
					}
					if msg.Action == ActionStart {
						started++
					}
				}
			case err := <-errC:
				fmt.Println(err)
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-interrupt

	cancel()

	fmt.Printf("\nStatistics from %s to %s\n", startedAt.Format(time.RFC3339), time.Now().Format(time.RFC3339))
	fmt.Printf("Created : %d\n", created)
	fmt.Printf("Started : %d\n", started)
}
