package main

import (
	"fmt"
	"github.com/BumpyClock/hermes"
)

func main() {
	client := hermes.New()
	fmt.Printf("Hermes client created: %T\n", client)
}
