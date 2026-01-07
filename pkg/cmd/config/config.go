package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/cylonchau/pantheon/pkg/api/config"
)

var (
	configExample = templates.Examples(i18n.T(`
		# initial config file
		pantheonctl config init
	`))
)

// NewCmdConfig creates the config command and adds child commands
func NewCmdConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Pantheon CLI configuration",
	}

	// Add subcommands
	cmd.AddCommand(newCmdConfigInit())
	cmd.AddCommand(newCmdAddCluster())
	cmd.AddCommand(newCmdDeleteCluster())
	cmd.AddCommand(newCmdListClusters())
	cmd.AddCommand(newCmdSetContext())
	cmd.AddCommand(newCmdGetContext())

	return cmd
}

// GetConfigPath retrieves the config file path from the environment or uses the default
func GetConfigPath() string {
	if path := os.Getenv("PANTHEONCONFIG"); path != "" {
		return path
	}
	return os.ExpandEnv("$HOME/.pantheon")
}

func readConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &config.Config{}, nil // If config does not exist, return empty config
		}
		return nil, err
	}

	var config config.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}
	return &config, nil
}

// writeConfig writes the configuration back to the file
func writeConfig(path string, config *config.Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// GetClusterConfig retrieves the cluster configuration for the current context
func GetClusterConfig() (*config.ClusterConfig, error) {
	file := GetConfigPath() + "/" + "config"
	config, err := readConfig(file)
	if err != nil {
		return nil, err
	}

	// Find the cluster that matches the current context
	for _, cluster := range config.Clusters {
		if cluster.Name == config.CurrentContext {
			return &cluster, nil
		}
	}

	return nil, fmt.Errorf("no cluster found for current context: %s", config.CurrentContext)
}
