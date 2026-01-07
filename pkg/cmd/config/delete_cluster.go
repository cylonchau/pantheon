package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cylonchau/pantheon/pkg/api/config"
)

// NewCmdDeleteCluster creates the delete-cluster command
func newCmdDeleteCluster() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "delete-cluster",
		Short: "Delete a cluster from the configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return DeleteCluster(name)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name of the cluster to delete (required)")
	cmd.MarkFlagRequired("name")

	return cmd
}

// DeleteCluster removes a cluster from the configuration file
func DeleteCluster(name string) error {
	file := GetConfigPath() + "config"
	configFile, err := readConfig(file)
	if err != nil {
		return err
	}

	var updatedClusters []config.ClusterConfig
	for _, cluster := range configFile.Clusters {
		if cluster.Name != name {
			updatedClusters = append(updatedClusters, cluster)
		}
	}

	if len(updatedClusters) == len(configFile.Clusters) {
		return fmt.Errorf("cluster with name %s not found", name)
	}

	configFile.Clusters = updatedClusters
	return writeConfig(file, configFile)
}
