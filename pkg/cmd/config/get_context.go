package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewCmdGetContext creates the get-context command
func newCmdGetContext() *cobra.Command {
	return &cobra.Command{
		Use:   "get-context",
		Short: "Get the current context",
		RunE: func(cmd *cobra.Command, args []string) error {
			return getContext()
		},
	}
}

// GetContext retrieves the current context from the configuration
func getContext() error {
	file := GetConfigPath() + "/" + "config"
	config, err := readConfig(file)
	if err != nil {
		return err
	}

	// Check if current context is set
	if config.CurrentContext == "" {
		return fmt.Errorf("no current context set")
	}

	// Print the current context
	fmt.Printf("Current context: %s\n", config.CurrentContext)

	return nil
}
