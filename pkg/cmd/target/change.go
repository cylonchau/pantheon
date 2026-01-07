package target

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/cylonchau/pantheon/pkg/api/target"
	"github.com/cylonchau/pantheon/pkg/cmd/config"
	"github.com/cylonchau/pantheon/pkg/cmd/path_map"
	"github.com/cylonchau/pantheon/pkg/utils"
)

var (
	targetChangeExample = templates.Examples(i18n.T(`
		# Change a target's configuration
		pantheonctl target change --address 127.0.0.1:9090  --id 1

		# short cmd
		pantheonctl target chg --address 127.0.0.1:9090 --id 1

		# Change a target with labels
		pantheonctl target change --metric-path /prometheus --id 1
	`))
)

type TargetChangeOptions struct {
	ID            int
	Address       string
	MetricPath    string
	ScrapeTime    int
	ScrapeTimeout int
	Auth          TargetAuth
}

// NewTargetChangeOptions creates the options for changing a target with default values
func newTargetChangeOptions() *TargetChangeOptions {
	return &TargetChangeOptions{
		Auth: TargetAuth{},
	}
}

// NewCmdTargetChange creates a new Change command.
func newCmdTargetChange() *cobra.Command {
	o := newTargetChangeOptions()

	changeCmd := &cobra.Command{
		Use:     "change --address=127.0.0.1:9090 --selector prom=fed",
		Short:   i18n.T("Change a target"),
		Example: targetChangeExample,
		Aliases: []string{"chg"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(cmd); err != nil {
				return err
			}
			if err := o.Validate(cmd, args); err != nil {
				return err
			}
			return o.Run(args)
		},
	}

	// Define flags with default values
	changeCmd.Flags().IntVar(&o.ID, "id", 0, "Specify the target id of the target to change.")
	changeCmd.Flags().StringVar(&o.Address, "address", o.Address, "Specify the address of the target. This is required.")
	changeCmd.Flags().StringVar(&o.MetricPath, "metric-path", "", "Specify the metric path of the target.")
	changeCmd.Flags().IntVar(&o.ScrapeTime, "scrape-time", 0, "Specify the scrape time of the target.")
	changeCmd.Flags().IntVar(&o.ScrapeTimeout, "scrape-timeout", 0, "Specify the scrape timeout of the target.")
	changeCmd.Flags().StringVar(&o.Auth.Base, "auth-base", "", "Specify the base auth of the target. This is optional.")
	changeCmd.Flags().StringVar(&o.Auth.BearerToken, "auth-bearer", "", "Specify the bearer token of the target. This is optional.")
	changeCmd.MarkFlagRequired("id")
	return changeCmd
}

// Complete processes the command line arguments and populates the options
func (o *TargetChangeOptions) Complete(cmd *cobra.Command) error {
	if err := o.processFlags(); err != nil {
		return err
	}
	return nil
}

func (o *TargetChangeOptions) processFlags() error {
	return nil
}

func (o *TargetChangeOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *TargetChangeOptions) Run(args []string) error {
	cluster, err := config.GetClusterConfig()
	if err != nil {
		return err
	}

	targetQuery := target.TargetChg{
		Address:       o.Address,
		MetricPath:    o.MetricPath,
		ScrapeTime:    o.ScrapeTime,
		ScrapeTimeout: o.ScrapeTimeout,
		Auth: &target.TargetAuth{
			Base:        o.Auth.Base,
			BearerToken: o.Auth.BearerToken,
		},
	}

	body, err := json.Marshal(targetQuery)
	if err != nil {
		return err
	}

	api, exists := path_map.APIInterfaces["ChangeTarget"]
	if !exists {
		return fmt.Errorf("unsupported API")
	}
	url := fmt.Sprintf("%s%s/%d", cluster.Cluster.Server, api.Path, o.ID)

	resp, err := utils.SendRequest(api.Method, url, body, cluster.Cluster.Auth)
	if err != nil {
		return err
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

		return fmt.Errorf("failed to change target: %s", responseBody.Msg)
	}

	fmt.Printf("target <%d> updated\n", o.ID)
	return nil
}
