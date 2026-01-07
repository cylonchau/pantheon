package target

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

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
	targetAddExample = templates.Examples(i18n.T(`
		# Add a target
		pantheonctl target add --address 127.0.0.1:9090 --selector prom=fed

		# Add a target with labels
		pantheonctl target add --address 127.0.0.1:9090 --labels dc=prd-190 --selector prom=fed
		
		# To attach multiple labels to the target.
		pantheonctl target add --address 127.0.0.1:9090 --labels dc=prd-190,module=a04-ws --selector prom=fed

		# To attach multiple param to the target (for blackbox).
		pantheonctl target add --address localhost:9115 --labels dc=prd-190,module=a04-ws --params target=google.com,module=http_2xx --selector prom=fed
	`))
)

type TargetAddOptions struct {
	Address         string        `form:"address" json:"address" yaml:"address" binding:"required"`
	MetricPath      string        `form:"metric_path,default=/metrics" json:"metric_path,default=/metrics" yaml:"metric_path"`
	ScrapeTime      int           `form:"scrape_time,default=30" json:"scrape_time,default=30" yaml:"scrape_time"`
	ScrapeTimeout   int           `form:"scrape_timeout,default=10" json:"scrape_timeout,default=10" yaml:"scrape_timeout"`
	Labels          []TargetLabel `json:"labels,omitempty" yaml:"labels,omitempty" form:"labels,omitempty"`
	Selectors       []TargetLabel `json:"instanceSelector"` // 确保这是必需的
	Params          []TargetLabel `json:"params"`           // 确保这是必需的
	LabelsString    string
	SelectorsString string
	ParamsString    string
	Auth            TargetAuth
}

// NewTargetOptions creates the options for target with default values
func newTargetAddOptions() *TargetAddOptions {
	return &TargetAddOptions{
		Address:       "",         // Address is required, so no default value here
		MetricPath:    "/metrics", // Default metric path
		ScrapeTime:    30,         // Default scrape time is 30s
		ScrapeTimeout: 10,         // Default timeout is 10s
		Labels:        []TargetLabel{},
		Selectors:     []TargetLabel{},
		Params:        []TargetLabel{},
		Auth:          TargetAuth{},
	}
}

// NewCmdTarget creates a new Target command.
func newCmdTargetAdd() *cobra.Command {
	o := newTargetAddOptions()

	addCmd := &cobra.Command{
		Use:     "add --address=127.0.0.1:9090 --selector prom=fed",
		Short:   i18n.T("Add a target"),
		Aliases: []string{"create"},
		Example: targetAddExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Complete and validate will be run before the main logic
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
	addCmd.Flags().StringVar(&o.Address, "address", o.Address, "Specify the address of the target. This is required.")
	addCmd.Flags().StringVar(&o.MetricPath, "metric-path", "/metrics", "Specify the metric path of the target.")
	addCmd.Flags().IntVar(&o.ScrapeTime, "scrape-time", 30, "Specify the scrape time of the target.")
	addCmd.Flags().IntVar(&o.ScrapeTimeout, "scrape-timeout", 10, "Specify the scrape timeout of the target.")
	addCmd.Flags().StringVar(&o.LabelsString, "labels", "", "Comma-separated key=value pairs for labels.")
	addCmd.Flags().StringVar(&o.SelectorsString, "selector", "", "Comma-separated key=value pairs for instance selectors. This is required.")
	addCmd.Flags().StringVar(&o.ParamsString, "params", "", "Comma-separated key=value pairs for target paramters. This is optional.")
	addCmd.Flags().StringVar(&o.Auth.Base, "auth-base", "", "Specify the base auth of the target. This is optional.")
	addCmd.Flags().StringVar(&o.Auth.BearerToken, "auth-bearer", "", "Specify the bearer token of the target. This is optional.")
	addCmd.MarkFlagRequired("address")
	addCmd.MarkFlagRequired("selector")
	return addCmd
}

// Complete processes the command line arguments and populates the options
func (o *TargetAddOptions) Complete(cmd *cobra.Command) error {
	// Validate and process labels
	if o.LabelsString != "" {
		pairs := strings.Split(o.LabelsString, ",")
		for _, pair := range pairs {
			kv := strings.Split(pair, "=")
			if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
				return fmt.Errorf("invalid format for labels: expected 'key=value' or 'key1=value1,key2=value2'")
			}
			o.Labels = append(o.Labels, TargetLabel{Key: kv[0], Value: kv[1]})
		}
	}

	// Validate and process params
	if o.ParamsString != "" {
		pairs := strings.Split(o.ParamsString, ",")
		for _, pair := range pairs {
			kv := strings.Split(pair, "=")
			if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
				return fmt.Errorf("invalid format for params: expected 'key=value' or 'key1=value1,key2=value2'")
			}
			o.Params = append(o.Params, TargetLabel{Key: kv[0], Value: kv[1]})
		}
	}

	// Validate and process selectors
	if o.SelectorsString != "" {
		pairs := strings.Split(o.SelectorsString, ",")
		for _, pair := range pairs {
			kv := strings.Split(pair, "=")
			if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
				return fmt.Errorf("invalid format for selector: expected 'key=value' or 'key1=value1,key2=value2'")
			}
			o.Selectors = append(o.Selectors, TargetLabel{Key: kv[0], Value: kv[1]})
		}
	}
	return nil
}

func (o *TargetAddOptions) Validate(cmd *cobra.Command, args []string) error {
	// Define pattern for valid keys (letters, numbers, hyphen, starting with a letter)
	keyPattern := "^[a-zA-Z][a-zA-Z0-9-]*$"

	// Validate labels and selectors for key=value format and proper key
	if err := validateKeyValuePairs(o.Labels, keyPattern); err != nil {
		return err
	}
	if err := validateKeyValuePairs(o.Selectors, keyPattern); err != nil {
		return err
	}

	return nil
}

func (o *TargetAddOptions) Run(args []string) error {
	// Get the server and tokens from config
	cluster, err := config.GetClusterConfig()
	if err != nil {
		return err
	}

	// Prepare the request body
	targetQuery := target.Target{
		Targets: []target.TargetItem{
			{
				Address:       o.Address,
				MetricPath:    o.MetricPath,
				ScrapeTime:    o.ScrapeTime,
				ScrapeTimeout: o.ScrapeTimeout,
				Auth: &target.TargetAuth{
					Base:        o.Auth.Base,
					BearerToken: o.Auth.BearerToken,
				},
				Labels: convertToRequestType(o.Labels),
				Params: convertToRequestType(o.Params),
			},
		},
		InstanceSelector: convertToRequestType(o.Selectors),
	}

	body, err := json.Marshal(targetQuery)
	if err != nil {
		return err
	}

	// Get the AddTarget API path using the map
	api, exists := path_map.APIInterfaces["AddTarget"]
	if !exists {
		return fmt.Errorf("Unsupport API ")
	}
	url := fmt.Sprintf("%s%s", cluster.Cluster.Server, api.Path)

	// Send the HTTP request
	resp, err := utils.SendRequest(api.Method, url, body, cluster.Cluster.Auth) // Directly pass the Auth info
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add target, received status: %s", resp.Status)
	}

	if resp.StatusCode != http.StatusOK {
		// 读取响应体
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		// 使用 sonic 解析响应体
		var responseBody struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}

		if err := sonic.Unmarshal(body, &responseBody); err != nil {
			return fmt.Errorf("failed to decode response body: %w", err)
		}

		return fmt.Errorf("failed to add target: %s", responseBody.Msg)
	}

	fmt.Printf("target <%s> created\n", o.Address)
	return nil
}
