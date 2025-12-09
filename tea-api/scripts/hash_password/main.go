package main

import (
	"fmt"
	"os"
	"tea-api/pkg/utils"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./scripts/hash_password <password>")
		return
	}
	pw := os.Args[1]
	h, err := utils.HashPassword(pw)
	if err != nil {
		fmt.Println("hash error:", err)
		return
	}
	fmt.Println(h)
}
