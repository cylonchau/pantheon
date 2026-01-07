package config

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// NewCmdListClusters creates the list-clusters command
func newCmdListClusters() *cobra.Command {
	return &cobra.Command{
		Use:   "list-clusters",
		Short: "List all clusters in the configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ListClusters()
		},
	}
}

// ListClusters prints all clusters in the configuration
func ListClusters() error {
	file := GetConfigPath() + "/" + "config"
	config, err := readConfig(file)
	if err != nil {
		return err
	}

	if len(config.Clusters) == 0 {
		return fmt.Errorf("no cluster found.")
	}

	// 使用 tabwriter 创建输出
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// 打印表头
	fmt.Fprintf(w, "Cluster\tServer\tAuth Type\n")

	// 打印集群详细信息
	for _, cluster := range config.Clusters {
		authType := "none"
		if cluster.Cluster.Auth.BaseAuth != "" {
			authType = "BaseAuth"
		} else if cluster.Cluster.Auth.BearerToken != "" {
			authType = "BearerToken"
		} else if cluster.Cluster.Auth.SSOToken != "" {
			authType = "SSOToken"
		}
		// 打印行
		fmt.Fprintf(w, "%s\t%s\t%s\n", cluster.Name, cluster.Cluster.Server, authType)
	}

	// 刷新并输出表格
	w.Flush()
	return nil
}
