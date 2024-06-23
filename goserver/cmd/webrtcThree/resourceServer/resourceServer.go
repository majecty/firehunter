package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	fmt.Println("resourceServer start")
	ctx := context.Background()
	if err := webrtcMain(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func webrtcMain(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	fmt.Println("webrtcMain start")

	<-ctx.Done()
	fmt.Println("webrtcMain end")
	return nil
}
