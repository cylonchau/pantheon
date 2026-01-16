package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cylonchau/pantheon/pkg/version"
)

func NewCmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of pantheonctl",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Pantheon version: %s\n", version.Version)
		},
	}
}
