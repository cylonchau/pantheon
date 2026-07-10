package monitor

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/cylonchau/pantheon/pkg/api/monitor"
	"github.com/cylonchau/pantheon/pkg/cmd/config"
	"github.com/cylonchau/pantheon/pkg/cmd/path_map"
	"github.com/cylonchau/pantheon/pkg/utils"
)

var (
	addExample = templates.Examples(i18n.T(`
		# Add a new pod monitor rule
		pantheonctl monitor add --name=my-pod-rule --type=pod --namespace=default --selector=app=test-app --port=metrics`))
)

type monitorAddOptions struct {
	name        string
	ruleType    string
	namespace   string
	selector    string
	portName    string
	metricPath  string
	labels      string
	dropMetrics string
}

func newCmdMonitorAdd() *cobra.Command {
	o := &monitorAddOptions{}

	addCmd := &cobra.Command{
		Use:     "add",
		Short:   i18n.T("Add a new monitor rule"),
		Example: addExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	addCmd.Flags().StringVar(&o.name, "name", "", "Name of the monitor rule")
	addCmd.Flags().StringVar(&o.ruleType, "type", "", "Type of monitor target: pod or service")
	addCmd.Flags().StringVar(&o.namespace, "namespace", "", "Target Kubernetes namespace")
	addCmd.Flags().StringVar(&o.selector, "selector", "", "Label selector, e.g. app=test-app")
	addCmd.Flags().StringVar(&o.portName, "port", "", "Name or number of target port")
	addCmd.Flags().StringVar(&o.metricPath, "metric-path", "/metrics", "HTTP path to scrape metrics from")
	addCmd.Flags().StringVar(&o.labels, "labels", "", "Key-value labels to inject, e.g. env=prod,owner=team")
	addCmd.Flags().StringVar(&o.dropMetrics, "drop-metrics", "", "Regex to filter out unwanted metrics")

	addCmd.MarkFlagRequired("name")
	addCmd.MarkFlagRequired("type")
	addCmd.MarkFlagRequired("namespace")
	addCmd.MarkFlagRequired("selector")
	addCmd.MarkFlagRequired("port")

	return addCmd
}

func (o *monitorAddOptions) Run() error {
	if o.ruleType != "pod" && o.ruleType != "service" {
		return fmt.Errorf("invalid type: %s. Must be 'pod' or 'service'", o.ruleType)
	}

	cluster, err := config.GetClusterConfig()
	if err != nil {
		return err
	}
	api, exists := path_map.APIInterfaces["AddMonitor"]
	if !exists {
		return fmt.Errorf("Unsupported API")
	}

	url := fmt.Sprintf("%s%s", cluster.Cluster.Server, api.Path)

	reqPayload := monitor.MonitorRuleReq{
		Name:           o.name,
		Type:           o.ruleType,
		Namespace:      o.namespace,
		SelectorString: o.selector,
		PortName:       o.portName,
		MetricPath:     o.metricPath,
		LabelsString:   o.labels,
		DropMetrics:    o.dropMetrics,
	}

	reqBytes, err := sonic.Marshal(reqPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal request payload: %w", err)
	}

	resp, err := utils.SendRequest(api.Method, url, reqBytes, cluster.Cluster.Auth)
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

	fmt.Println("Monitor rule added successfully.")
	return nil
}
