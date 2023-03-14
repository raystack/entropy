package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/goto/entropy/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cli.Execute(ctx)
}
