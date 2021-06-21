package main

import "github.com/jacklaaa89/pokeapi/cmd"

// main is the main entrypoint to the application.
// its only function is to execute the root command.
func main() {
	cmd.Root().Execute()
}
