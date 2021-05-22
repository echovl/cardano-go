package main

import "github.com/echovl/cardano-wallet/cmd"

func main() {
	defer cmd.DefaultDb.Close()

	cmd.Execute()
}
