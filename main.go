package main

import (
	"os"

	"github.com/AlonzaT/blockchain/cli_client"
)

func main() {
	defer os.Exit(0)
	cli := cli_client.CommandLine{}

	cli.Run()

}
