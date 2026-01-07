package target

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/bytedance/sonic"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/cylonchau/pantheon/pkg/api/target"
	"github.com/cylonchau/pantheon/pkg/cmd/config"
	"github.com/cylonchau/pantheon/pkg/cmd/path_map"
	"github.com/cylonchau/pantheon/pkg/utils"
)

// 定义列表命令的使用示例
var (
	listExample = templates.Examples(i18n.T(`
		# List all targets with specific labels
		pantheonctl target list --selector dc=prd-190
		
		# short
        pantheonctl target ls --selector dc=prd-190`))
)

// TargetListOptions holds the options for the list command
type TargetListOptions struct {
	SelectorString string
	Selector       []TargetLabel
	IsShowLabels   bool
	IsShowParams   bool
	OutputFormat   string
}

// NewTargetListOptions creates the options for the list command
func NewTargetListOptions() *TargetListOptions {
	return &TargetListOptions{}
}

// NewCmdTargetList creates a new list command
func newCmdTargetList() *cobra.Command {
	o := NewTargetListOptions()

	listCmd := &cobra.Command{
		Use:     "list",
		Short:   i18n.T("List targets"),
		Aliases: []string{"ls"},
		Example: listExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Complete and validate will be run before the main logic
			if err := o.Complete(cmd); err != nil {
				return err
			}
			if err := o.Validate(cmd, args); err != nil {
				return err
			}
			return o.Run()
		},
	}

	// 定义 flags
	listCmd.Flags().StringVar(&o.SelectorString, "selector", "", "Comma-separated key=value pairs for selectors (required)")
	listCmd.Flags().BoolVar(&o.IsShowLabels, "show-labels", false, "When printing, show all labels as the last column (default hide labels column)")
	listCmd.Flags().BoolVar(&o.IsShowParams, "show-params", false, "When printing, show all parameters as the last column (default hide parameters column)")
	listCmd.Flags().StringVarP(&o.OutputFormat, "output", "o", "", "Output format. One of: json|yaml")
	listCmd.MarkFlagRequired("selector")
	return listCmd
}

// Complete processes the labels argument
func (o *TargetListOptions) Complete(cmd *cobra.Command) error {
	if o.SelectorString != "" {
		pairs := strings.Split(o.SelectorString, ",")
		for _, pair := range pairs {
			kv := strings.Split(pair, "=")
			if len(kv) == 2 {
				o.Selector = append(o.Selector, TargetLabel{Key: kv[0], Value: kv[1]})
			} else {
				return fmt.Errorf("Invalid selector format: %s. Expected format: key=value", pair)
			}
		}
	}
	return nil
}

// Validate ensures the required arguments and flags are provided and valid
func (o *TargetListOptions) Validate(cmd *cobra.Command, args []string) error {
	// Validate the output format
	if o.OutputFormat != "" {
		if o.OutputFormat != "json" && o.OutputFormat != "yaml" {
			return fmt.Errorf("invalid output format: %s. Valid values are 'json' or 'yaml'", o.OutputFormat)
		}
	}
	return nil
}

// Run lists the targets
func (o *TargetListOptions) Run() error {
	// 假设这里请求 HTTP API 返回目标列表
	targets, err := o.listTargetsFromAPI()
	if err != nil {
		return err
	}

	// 根据用户指定的输出格式输出结果
	switch o.OutputFormat {
	case "json":
		return printJSON(targets)
	case "yaml":
		return printYAML(targets)
	default:
		return printTable(targets, o.IsShowLabels, o.IsShowParams)
	}
}

func (o *TargetListOptions) listTargetsFromAPI() ([]target.TargetList, error) {

	cluster, err := config.GetClusterConfig()
	if err != nil {
		return nil, err
	}
	api, exists := path_map.APIInterfaces["ListCmdTargets"]
	if !exists {
		return nil, fmt.Errorf("Unsupport API")
	}

	// 构建 URL，假设 selectors 至少包含一个值
	url := fmt.Sprintf("%s%s/%s/%s", cluster.Cluster.Server, api.Path, o.Selector[0].Key, o.Selector[0].Value)

	// 发送 HTTP 请求
	resp, err := utils.SendRequest(api.Method, url, nil, cluster.Cluster.Auth)
	if err != nil {
		return nil, err
	}
	// defer resp.Body.Close() 应该放在读取内容之后
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list targets, received status: %s", resp.Status)
	}

	// 使用 io.ReadAll 读取 resp.Body
	body, err := io.ReadAll(resp.Body) // 先读取所有内容，再 defer 关闭
	if err != nil {
		return nil, err
	}

	// 使用 Sonic 解析 JSON 响应
	var targets []target.TargetList
	err = sonic.Unmarshal(body, &targets)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response using sonic: %w", err)
	}

	return targets, nil
}

func printTable(targets []target.TargetList, showLabels bool, showParams bool) error {

	// 检查 targets 是否为空
	if len(targets) == 0 {
		fmt.Println("No resources found.")
		return nil
	}

	// 预先计算每一列的最大宽度，基于目标列表的内容
	maxIDWidth := len("ID")
	maxAddressWidth := len("ADDRESS")
	maxMetricPathWidth := len("METRIC_PATH")
	maxScrapeTimeWidth := len("SCRAPE_TIME")
	maxScrapeTimeoutWidth := len("SCRAPE_TIMEOUT")
	maxAuthTypeWidth := len("AUTH_TYPE")
	maxLabelsWidth := len("LABELS")
	maxParamsWidth := len("PARAMETERS")

	// 遍历每个 target，计算每列中最长的宽度
	for _, target := range targets {
		if len(fmt.Sprintf("%d", target.ID)) > maxIDWidth {
			maxIDWidth = len(fmt.Sprintf("%d", target.ID))
		}
		if len(target.Address) > maxAddressWidth {
			maxAddressWidth = len(target.Address)
		}
		if len(target.MetricPath) > maxMetricPathWidth {
			maxMetricPathWidth = len(target.MetricPath)
		}
		scrapeTimeStr := fmt.Sprintf("%d", target.ScrapeTime)
		if len(scrapeTimeStr) > maxScrapeTimeWidth {
			maxScrapeTimeWidth = len(scrapeTimeStr)
		}

		authType := "None"
		if target.Auth != nil {
			if target.Auth.Base != "" {
				authType = "Base Auth"
			} else if target.Auth.BearerToken != "" {
				authType = "Bearer Token"
			}
		}
		if len(authType) > maxAuthTypeWidth {
			maxAuthTypeWidth = len(authType)
		}

		if showLabels {
			labels := []string{}
			for key, value := range target.Labels {
				labels = append(labels, fmt.Sprintf("%s=%s", key, value))
			}
			labelsStr := strings.Join(labels, ",")
			if len(labelsStr) > maxLabelsWidth {
				maxLabelsWidth = len(labelsStr)
			}
		}

		if showParams {
			params := []string{}
			for key, value := range target.Params {
				params = append(params, fmt.Sprintf("%s=%s", key, value))
			}
			paramsStr := strings.Join(params, ",")
			if len(paramsStr) > maxParamsWidth {
				maxParamsWidth = len(paramsStr)
			}
		}
	}

	// 设置格式化字符串，动态控制列宽
	var headerFormat string

	if showLabels && showParams {
		headerFormat = fmt.Sprintf("%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds", maxIDWidth, maxAddressWidth, maxMetricPathWidth, maxScrapeTimeWidth, maxScrapeTimeoutWidth, maxAuthTypeWidth, maxLabelsWidth, maxParamsWidth)
	} else if showLabels {
		headerFormat = fmt.Sprintf("%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds", maxIDWidth, maxAddressWidth, maxMetricPathWidth, maxScrapeTimeWidth, maxScrapeTimeoutWidth, maxAuthTypeWidth, maxLabelsWidth)
	} else if showParams {
		headerFormat = fmt.Sprintf("%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds", maxIDWidth, maxAddressWidth, maxMetricPathWidth, maxScrapeTimeWidth, maxScrapeTimeoutWidth, maxAuthTypeWidth, maxParamsWidth)
	} else {
		headerFormat = fmt.Sprintf("%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds\t%%-%ds", maxIDWidth, maxAddressWidth, maxMetricPathWidth, maxScrapeTimeWidth, maxScrapeTimeoutWidth, maxAuthTypeWidth)
	}

	var rowFormat string
	if showLabels && showParams {
		rowFormat = fmt.Sprintf("%%-%dd\t%%-%ds\t%%-%ds\t%%-%dd\t%%-%dd\t%%-%ds\t%%-%ds\t%%-%ds", maxIDWidth, maxAddressWidth, maxMetricPathWidth, maxScrapeTimeWidth, maxScrapeTimeoutWidth, maxAuthTypeWidth, maxLabelsWidth, maxParamsWidth)
	} else if showLabels {
		rowFormat = fmt.Sprintf("%%-%dd\t%%-%ds\t%%-%ds\t%%-%dd\t%%-%dd\t%%-%ds\t%%-%ds", maxIDWidth, maxAddressWidth, maxMetricPathWidth, maxScrapeTimeWidth, maxScrapeTimeoutWidth, maxAuthTypeWidth, maxLabelsWidth)
	} else if showParams {
		rowFormat = fmt.Sprintf("%%-%dd\t%%-%ds\t%%-%ds\t%%-%dd\t%%-%dd\t%%-%ds\t%%-%ds", maxIDWidth, maxAddressWidth, maxMetricPathWidth, maxScrapeTimeWidth, maxScrapeTimeoutWidth, maxAuthTypeWidth, maxParamsWidth)
	} else {
		rowFormat = fmt.Sprintf("%%-%dd\t%%-%ds\t%%-%ds\t%%-%dd\t%%-%dd\t%%-%ds", maxIDWidth, maxAddressWidth, maxMetricPathWidth, maxScrapeTimeWidth, maxScrapeTimeoutWidth, maxAuthTypeWidth)
	}

	// 使用 tabwriter 自动计算列宽，设置适合的填充和对齐
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// 打印表头
	if showLabels && showParams {
		fmt.Fprintf(w, headerFormat, "ID", "ADDRESS", "METRIC_PATH", "SCRAPE_TIME", "SCRAPE_TIMEOUT", "AUTH_TYPE", "LABELS", "PARAMETERS")
	} else if showLabels {
		fmt.Fprintf(w, headerFormat, "ID", "ADDRESS", "METRIC_PATH", "SCRAPE_TIME", "SCRAPE_TIMEOUT", "AUTH_TYPE", "LABELS")
	} else if showParams {
		fmt.Fprintf(w, headerFormat, "ID", "ADDRESS", "METRIC_PATH", "SCRAPE_TIME", "SCRAPE_TIMEOUT", "AUTH_TYPE", "PARAMETERS")
	} else {
		fmt.Fprintf(w, headerFormat, "ID", "ADDRESS", "METRIC_PATH", "SCRAPE_TIME", "SCRAPE_TIMEOUT", "AUTH_TYPE")
	}
	fmt.Fprintln(w) // 换行

	// 遍历目标列表并打印每一行
	for _, target := range targets {
		authType := "None"
		if target.Auth != nil {
			if target.Auth.Base != "" {
				authType = "Base Auth"
			} else if target.Auth.BearerToken != "" {
				authType = "Bearer Token"
			}
		}

		// 如果显示 labels，则打印 labels 列
		if showLabels && showParams {
			labels := []string{}
			for key, value := range target.Labels {
				labels = append(labels, fmt.Sprintf("%s=%s", key, value))
			}
			params := []string{}
			for key, value := range target.Params {
				params = append(params, fmt.Sprintf("%s=%s", key, value))
			}

			// 打印目标的基本信息
			fmt.Fprintf(w, rowFormat,
				target.ID,
				target.Address,
				target.MetricPath,
				target.ScrapeTime,
				target.ScrapeTimeout,
				authType,
				strings.Join(labels, ","),
				strings.Join(params, ","),
			)
			fmt.Fprintln(w)
		} else if showLabels {
			labels := []string{}
			for key, value := range target.Labels {
				labels = append(labels, fmt.Sprintf("%s=%s", key, value))
			}
			fmt.Fprintf(w, rowFormat, target.ID, target.Address, target.MetricPath, target.ScrapeTime, target.ScrapeTimeout, authType, strings.Join(labels, ","))
			fmt.Fprintln(w)
		} else if showParams {
			params := []string{}
			for key, value := range target.Params {
				params = append(params, fmt.Sprintf("%s=%s", key, value))
			}
			fmt.Fprintf(w, rowFormat, target.ID, target.Address, target.MetricPath, target.ScrapeTime, target.ScrapeTimeout, authType, strings.Join(params, ","))
			fmt.Fprintln(w)
		} else {
			fmt.Fprintf(w, rowFormat, target.ID, target.Address, target.MetricPath, target.ScrapeTime, target.ScrapeTimeout, authType)
			fmt.Fprintln(w)
		}
	}

	// 刷新并输出表格
	w.Flush()
	return nil
}

// 打印 JSON 格式输出
func printJSON(targets []target.TargetList) error {
	data, err := json.MarshalIndent(targets, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// 打印 YAML 格式输出
func printYAML(targets []target.TargetList) error {
	data, err := yaml.Marshal(targets)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
