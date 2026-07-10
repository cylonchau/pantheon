package monitor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/bytedance/sonic"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/cylonchau/pantheon/pkg/cmd/config"
	"github.com/cylonchau/pantheon/pkg/cmd/path_map"
	"github.com/cylonchau/pantheon/pkg/model"
	"github.com/cylonchau/pantheon/pkg/utils"
)

var (
	listExample = templates.Examples(i18n.T(`
		# List all active monitor rules
		pantheonctl monitor list`))
)

type monitorListOptions struct{}

func newCmdMonitorList() *cobra.Command {
	o := &monitorListOptions{}

	listCmd := &cobra.Command{
		Use:     "list",
		Short:   i18n.T("List active monitor rules"),
		Aliases: []string{"ls"},
		Example: listExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}
	return listCmd
}

func (o *monitorListOptions) Run() error {
	rules, err := o.listMonitorsFromAPI()
	if err != nil {
		return err
	}

	if len(rules) == 0 {
		fmt.Println("No monitor rules found.")
		return nil
	}

	return printTable(rules)
}

func (o *monitorListOptions) listMonitorsFromAPI() ([]model.MonitorRule, error) {
	cluster, err := config.GetClusterConfig()
	if err != nil {
		return nil, err
	}
	api, exists := path_map.APIInterfaces["ListMonitors"]
	if !exists {
		return nil, fmt.Errorf("Unsupported API")
	}

	url := fmt.Sprintf("%s%s", cluster.Cluster.Server, api.Path)

	resp, err := utils.SendRequest(api.Method, url, nil, cluster.Cluster.Auth)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	var rules []model.MonitorRule
	if err := sonic.Unmarshal(body, &rules); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return rules, nil
}

func printTable(rules []model.MonitorRule) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTYPE\tNAMESPACE\tSELECTOR\tPORT_NAME\tMETRIC_PATH\tLABELS\tDROP_METRICS")
	for _, r := range rules {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			r.ID, r.Name, r.Type, r.Namespace, r.SelectorString, r.PortName, r.MetricPath, r.LabelsString, r.DropMetrics)
	}
	return w.Flush()
}
