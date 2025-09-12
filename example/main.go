package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Ajinx1/go-message-pattern-server/src/rpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := &rpc.Config{
		Addr:    ":4061",
		Timeout: 30 * time.Second,
	}

	server := rpc.NewServer(config)
	wrapper := rpc.NewServerWrapper(server)
	wrapper.MessagePattern("ping", func(data json.RawMessage) (interface{}, error) {
		return "pong", nil
	})

	go server.Start()
	fmt.Println("RPC Server started on :4061")

	time.Sleep(time.Second * 100)
	server.Shutdown(ctx)
}
