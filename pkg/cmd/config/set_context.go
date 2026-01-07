package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewCmdSetContext creates the set-context command
func newCmdSetContext() *cobra.Command {
	var contextName string

	cmd := &cobra.Command{
		Use:   "set-context [context-name]",
		Short: "Set the current context",
		Args:  cobra.ExactArgs(1), // Expect exactly one argument
		RunE: func(cmd *cobra.Command, args []string) error {
			contextName = args[0]
			return setContext(contextName)
		},
	}

	return cmd
}

// SetContext sets the current context in the configuration
func setContext(contextName string) error {
	file := GetConfigPath() + "/config"
	config, err := readConfig(file)
	if err != nil {
		return err
	}

	// Check if the provided context name exists in the clusters
	exists := false
	for _, cluster := range config.Clusters {
		if cluster.Name == contextName {
			exists = true
			break
		}
	}

	if !exists {
		return fmt.Errorf("context %s does not exist", contextName)
	}

	// Set the current context
	config.CurrentContext = contextName

	// Save the updated configuration
	if err := writeConfig(file, config); err != nil {
		return err
	}

	fmt.Printf("Current context set to: %s\n", contextName)
	return nil
}
