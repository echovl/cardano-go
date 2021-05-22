package main

import "github.com/echovl/cardano-go/cli/cmd"

func main() {
	defer cmd.DefaultDb.Close()

	cmd.Execute()
}
