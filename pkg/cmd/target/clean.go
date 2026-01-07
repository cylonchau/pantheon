package target

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/cylonchau/pantheon/pkg/cmd/config"
	"github.com/cylonchau/pantheon/pkg/cmd/path_map"
	"github.com/cylonchau/pantheon/pkg/utils"
)

var (
	cleanExample = templates.Examples(i18n.T(`
		# Clean all targets marked as deleted
		pantheonctl target clean
	`))
)

// TargetCleanOptions holds the options for the clean command
type TargetCleanOptions struct {
	Yes bool
}

// NewTargetCleanOptions creates the options for the clean command
func NewTargetCleanOptions() *TargetCleanOptions {
	return &TargetCleanOptions{}
}

// NewCmdTargetClean creates a new clean command
func newCmdTargetClean() *cobra.Command {
	o := NewTargetCleanOptions()
	cmd := &cobra.Command{
		Use:     "clean",
		Short:   i18n.T("Clean all targets marked as deleted"),
		Example: cleanExample,
		Aliases: []string{"clr", "cln"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}
	cmd.Flags().BoolVarP(&o.Yes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}

// Run executes the clean command
func (o *TargetCleanOptions) Run() error {
	if !o.Yes {
		fmt.Print("Are you sure you want to clean all targets marked as deleted? (y/n): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" {
			return fmt.Errorf("clean operation cancelled")
		}
	}

	cluster, err := config.GetClusterConfig()
	if err != nil {
		return err
	}

	api, exists := path_map.APIInterfaces["CleanDeletedTargets"]
	if !exists {
		return fmt.Errorf("Unsupported API")
	}

	url := fmt.Sprintf("%s%s", cluster.Cluster.Server, api.Path)
	resp, err := utils.SendRequest(api.Method, url, nil, cluster.Cluster.Auth)
	if err != nil {
		return fmt.Errorf("failed to send clean request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		var responseBody struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}

		if err := sonic.Unmarshal(body, &responseBody); err != nil {
			return fmt.Errorf("failed to decode response body: %w", err)
		}

		return fmt.Errorf("failed to clean targets: %s", responseBody.Msg)
	}

	fmt.Println("All targets marked as deleted have been cleaned successfully.")
	return nil
}
