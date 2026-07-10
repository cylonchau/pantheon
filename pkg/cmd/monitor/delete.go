package monitor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/cylonchau/pantheon/pkg/cmd/config"
	"github.com/cylonchau/pantheon/pkg/cmd/path_map"
	"github.com/cylonchau/pantheon/pkg/utils"
)

var (
	deleteExample = templates.Examples(i18n.T(`
		# Delete a monitor rule by ID
		pantheonctl monitor delete 1`))
)

type monitorDeleteOptions struct{}

func newCmdMonitorDelete() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:     "delete [ID]",
		Short:   i18n.T("Delete a monitor rule by ID"),
		Example: deleteExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid ID: %s. Must be an integer", args[0])
			}
			o := &monitorDeleteOptions{}
			return o.Run(uint(id))
		},
	}
	return deleteCmd
}

func (o *monitorDeleteOptions) Run(id uint) error {
	cluster, err := config.GetClusterConfig()
	if err != nil {
		return err
	}
	api, exists := path_map.APIInterfaces["DeleteMonitor"]
	if !exists {
		return fmt.Errorf("Unsupported API")
	}

	// Append the ID to the API path
	url := fmt.Sprintf("%s%s/%d", cluster.Cluster.Server, api.Path, id)

	resp, err := utils.SendRequest(api.Method, url, nil, cluster.Cluster.Auth)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var responseBody struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		if err := sonic.Unmarshal(body, &responseBody); err == nil {
			return fmt.Errorf("API error: %s (code %d)", responseBody.Msg, responseBody.Code)
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	fmt.Printf("Monitor rule with ID %d deleted successfully.\n", id)
	return nil
}
