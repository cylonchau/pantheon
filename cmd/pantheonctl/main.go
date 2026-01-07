package main

import (
	"os"

	"github.com/cylonchau/pantheon/pkg/cmd"
)

func main() {
	command := cmd.NewDefaultPantheonctlCommand()
	if err := command.Execute(); err != nil {
		// Pretty-print the error and exit with an error.
		//panic(err)
		os.Exit(199)
	}
}
