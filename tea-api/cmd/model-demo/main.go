package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"tea-api/pkg/modelclient"
)

func main() {
	prompt := flag.String("prompt", "测试：Hello", "prompt to send to model")
	flag.Parse()

	client, err := modelclient.NewAnthropicClientFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create model client: %v\n", err)
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	out, err := client.Generate(ctx, *prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generate error: %v\n", err)
		os.Exit(3)
	}

	fmt.Println("=== Model output ===")
	fmt.Println(out)
}
