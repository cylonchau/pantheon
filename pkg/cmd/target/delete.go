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
	deleteExample = templates.Examples(i18n.T(`
		# Delete a target
		pantheonctl target delete --id 1
		
		# short
        pantheonctl target rm --id 1`))
)

// TargetDeleteOptions holds the options for the delete command
type TargetDeleteOptions struct {
	ID     uint
	Yes    bool
	Labels []TargetLabel `json:"labels,omitempty" yaml:"labels,omitempty" form:"labels,omitempty"`
}

// NewTargetDeleteOptions creates the options for the delete command
func NewTargetDeleteOptions() *TargetDeleteOptions {
	return &TargetDeleteOptions{
		ID: 0,
	}
}

// NewCmdTargetDelete creates a new delete command
func newCmdTargetDelete() *cobra.Command {
	o := NewTargetDeleteOptions()
	cmd := &cobra.Command{
		Use:     "delete --address=127.0.0.1:9090 --id 1",
		Short:   i18n.T("Delete a target"),
		Aliases: []string{"rm", "remove", "del"},
		Example: deleteExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Complete and validate will be run before the main logic
			if err := o.Validate(cmd, args); err != nil {
				return err
			}
			return o.Run()
		},
	}
	// Define flags
	cmd.Flags().UintVar(&o.ID, "id", 0, "Specify the target id of the target to delete.")
	cmd.Flags().BoolVarP(&o.Yes, "yes", "y", false, "Skip confirmation prompt")
	cmd.MarkFlagRequired("id")
	return cmd
}

func (o *TargetDeleteOptions) Validate(cmd *cobra.Command, args []string) error {
	//// 验证第一个参数是否为整形
	//if _, err := strconv.Atoi(string(o.ID)); err != nil {
	//	return fmt.Errorf("ID argument '%s' is not a valid integer", o.ID)
	//}
	return nil
}

// Run deletes the target
func (o *TargetDeleteOptions) Run() error {

	// Get the server and tokens from config
	cluster, err := config.GetClusterConfig()
	if err != nil {
		return err
	}

	if o.ID > 0 {
		api, exists := path_map.APIInterfaces["GetTarget"]
		if !exists {
			return fmt.Errorf("Unsupported API")
		}
		url := fmt.Sprintf("%s%s/%d", cluster.Cluster.Server, api.Path, o.ID)
		// Send the HTTP request
		resp, err := utils.SendRequest(api.Method, url, nil, cluster.Cluster.Auth)
		if err != nil {
			return fmt.Errorf("failed to send delete request: %w", err)
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

			return fmt.Errorf("failed to delete target: %s", responseBody.Msg)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		var t map[string]interface{}
		err = sonic.Unmarshal(body, &t)
		if err != nil {
			return err
		}
		return confirmAndExecute(o.Yes, o.deleteTarget, fmt.Sprintf("Are you sure you want to delete target <%s://%s%s>?", t["schema"], t["address"], t["metric_path"]))
	}
	return fmt.Errorf("Invaild ID <%d>\n", o.ID)
}

// deleteTarget performs the deletion of the target
func (o *TargetDeleteOptions) deleteTarget() error {
	// Get the server and tokens from config
	cluster, err := config.GetClusterConfig()
	if err != nil {
		return err
	}

	if o.ID > 0 {
		api, exists := path_map.APIInterfaces["DeleteTargetWithID"]
		if !exists {
			return fmt.Errorf("Unsupported API")
		}
		url := fmt.Sprintf("%s%s/%d", cluster.Cluster.Server, api.Path, o.ID)
		// Send the HTTP request
		resp, err := utils.SendRequest(api.Method, url, nil, cluster.Cluster.Auth)
		if err != nil {
			return fmt.Errorf("failed to send delete request: %w", err)
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

			return fmt.Errorf("failed to delete target: %s", responseBody.Msg)
		}

		fmt.Printf("target <%d> deleted\n", o.ID)
	}
	return nil
}
