package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	config2 "github.com/cylonchau/pantheon/pkg/api/config"
)

// NewCmdConfigInit creates the init command
func newCmdConfigInit() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a blank configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitConfig()
		},
	}
}

// InitConfig creates an empty configuration file at the default location
func InitConfig() error {
	config := config2.Config{}
	data, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	path := GetConfigPath()
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	file := path + "/config"
	err = os.WriteFile(file, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	fmt.Println("Configuration file initialized:", file)
	return nil
}
