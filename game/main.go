package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	orbisplugin "example.com/mud/plugin"
	"example.com/mud/sdk"
)

func main() {
	debug := flag.Bool("debug", false, "run as a standalone gRPC server for IDE debugging")
	flag.Parse()

	if *debug {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		if err := orbisplugin.ServeDebug(ctx, sdk.NewAdapter(&Game{})); err != nil {
			log.Fatalf("debug serve: %v", err)
		}
		return
	}

	sdk.Serve(&Game{})
}
