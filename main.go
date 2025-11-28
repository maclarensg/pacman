package main

import (
	"fmt"
	"os"

	"pacman/game"
)

func main() {
	g, err := game.NewGame()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating game: %v\n", err)
		os.Exit(1)
	}

	defer g.Cleanup()

	g.Run()
}
