package selector

import (
	"fmt"
	"io"
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

// 定义列表命令的使用示例
var (
	listExample = templates.Examples(i18n.T(`
		# List all selectors
		pantheonctl selector list`))
)

// selectorListOptions holds the options for the list command
type selectorListOptions struct{}

// NewselectorListOptions creates the options for the list command
func NewselectorListOptions() *selectorListOptions {
	return &selectorListOptions{}
}

// NewCmdselectorList creates a new list command
func newCmdselectorList() *cobra.Command {
	o := NewselectorListOptions()

	listCmd := &cobra.Command{
		Use:     "list",
		Short:   i18n.T("List selectors"),
		Aliases: []string{"ls"},
		Example: listExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}
	return listCmd
}

// Run lists the selectors
func (o *selectorListOptions) Run() error {
	selectors, err := o.listselectorsFromAPI()
	if err != nil {
		return err
	}

	// 检查 selectors 是否为空
	if len(selectors) == 0 {
		fmt.Println("No resources found.")
		return nil
	}

	return printTable(selectors)
}

func (o *selectorListOptions) listselectorsFromAPI() ([]model.SelectorList, error) {
	cluster, err := config.GetClusterConfig()
	if err != nil {
		return nil, err
	}
	api, exists := path_map.APIInterfaces["ListCmdselectors"]
	if !exists {
		return nil, fmt.Errorf("Unsupported API")
	}

	url := fmt.Sprintf("%s%s", cluster.Cluster.Server, api.Path)

	resp, err := utils.SendRequest(api.Method, url, nil, cluster.Cluster.Auth)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 读取响应体
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		// 使用 sonic 解析响应体
		var responseBody struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}

		if err := sonic.Unmarshal(body, &responseBody); err != nil {
			return nil, fmt.Errorf("failed to decode response body: %w", err)
		}

		return nil, fmt.Errorf("failed to list selectors: %s", responseBody.Msg)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var selectors []model.SelectorList
	err = sonic.Unmarshal(body, &selectors)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response using sonic: %w", err)
	}

	return selectors, nil
}

func printTable(selectors []model.SelectorList) error {
	maxKeyWidth := len("KEY")
	maxValueWidth := len("VALUE")

	for _, selectorItem := range selectors {
		if len(selectorItem.Key) > maxKeyWidth {
			maxKeyWidth = len(selectorItem.Key)
		}
		if len(selectorItem.Value) > maxValueWidth {
			maxValueWidth = len(selectorItem.Value)
		}
	}

	headerFormat := fmt.Sprintf("%%-%ds\t%%-%ds\n", maxKeyWidth, maxValueWidth)
	rowFormat := fmt.Sprintf("%%-%ds\t%%-%ds\n", maxKeyWidth, maxValueWidth)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintf(w, headerFormat, "KEY", "VALUE")

	for _, selectorItem := range selectors {
		fmt.Fprintf(w, rowFormat, selectorItem.Key, selectorItem.Value)
	}

	w.Flush()
	return nil
}
