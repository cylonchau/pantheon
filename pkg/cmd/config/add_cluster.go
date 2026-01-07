package config

import (
	"github.com/spf13/cobra"

	"github.com/cylonchau/pantheon/pkg/api/config"
)

// NewCmdAddCluster creates the add-cluster command
func newCmdAddCluster() *cobra.Command {
	var name, server, baseAuth, bearerToken, ssoToken string

	cmd := &cobra.Command{
		Use:   "add-cluster",
		Short: "Add a new cluster to the configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return AddCluster(name, server, baseAuth, bearerToken, ssoToken)
		},
	}

	// Define flags
	cmd.Flags().StringVar(&name, "name", "", "Name of the cluster (required)")
	cmd.Flags().StringVar(&server, "server", "", "Server URL of the cluster (required)")
	cmd.Flags().StringVar(&baseAuth, "base-auth", "", "Base auth token")
	cmd.Flags().StringVar(&bearerToken, "bearer-token", "", "Bearer token")
	cmd.Flags().StringVar(&ssoToken, "sso-token", "", "SSO token")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("server")

	return cmd
}

// AddCluster adds a new cluster to the configuration file
func AddCluster(name, server, baseAuth, bearerToken, ssoToken string) error {
	file := GetConfigPath() + "/config"
	configFile, err := readConfig(file)
	if err != nil {
		return err
	}

	auth := config.Auth{
		BaseAuth:    baseAuth,
		BearerToken: bearerToken,
		SSOToken:    ssoToken,
	}

	cluster := config.ClusterConfig{
		Name: name,
		Cluster: config.Cluster{
			Server: server,
			Auth:   auth,
		},
	}

	configFile.Clusters = append(configFile.Clusters, cluster)
	return writeConfig(file, configFile)
}
