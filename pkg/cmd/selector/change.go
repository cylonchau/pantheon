package selector

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

// 定义变更命令的使用示例
var (
	changeExample = templates.Examples(i18n.T(`
		# Change a selector's key and value based on its old key and old value
		pantheonctl selector change --oldKey=myOldKey --oldValue=oldValue --newKey=myNewKey --newValue=newValue`))
)

// selectorChangeOptions holds the options for the change command
type selectorChangeOptions struct {
	oldKey   string
	oldValue string
	newKey   string
	newValue string
}

// NewselectorChangeOptions creates the options for the change command
func NewselectorChangeOptions() *selectorChangeOptions {
	return &selectorChangeOptions{}
}

// NewCmdselectorChange creates a new change command
func newCmdselectorChange() *cobra.Command {
	o := NewselectorChangeOptions()

	changeCmd := &cobra.Command{
		Use:     "change",
		Short:   i18n.T("Change a selector's key and value based on its old key and old value"),
		Aliases: []string{"chg"},
		Example: changeExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	// 添加 flags
	changeCmd.Flags().StringVarP(&o.oldKey, "oldKey", "k", "", "Current key of the selector to change")
	changeCmd.Flags().StringVarP(&o.oldValue, "oldValue", "v", "", "Current value of the selector")
	changeCmd.Flags().StringVar(&o.newKey, "newKey", "", "New key for the selector")
	changeCmd.Flags().StringVar(&o.newValue, "newValue", "", "New value for the selector (optional)")
	changeCmd.MarkFlagRequired("oldKey")
	changeCmd.MarkFlagRequired("oldValue")
	changeCmd.MarkFlagRequired("newKey")

	return changeCmd
}

// Run executes the change command
func (o *selectorChangeOptions) Run() error {
	return o.changeSelector()
}

// changeSelector handles the logic to change a selector's key and value based on old key and value
func (o *selectorChangeOptions) changeSelector() error {
	cluster, err := config.GetClusterConfig()
	if err != nil {
		return err
	}
	api, exists := path_map.APIInterfaces["ChangeCmdselectors"]
	if !exists {
		return fmt.Errorf("Unsupported API")
	}

	url := fmt.Sprintf("%s%s", cluster.Cluster.Server, api.Path)

	// 创建请求体
	body := map[string]string{
		"old_key":   o.oldKey,
		"old_value": o.oldValue,
		"new_key":   o.newKey,
		"new_value": o.newValue,
	}
	requestBody, err := sonic.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := utils.SendRequest(api.Method, url, requestBody, cluster.Cluster.Auth)
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

		return fmt.Errorf("failed to change selector: %s", responseBody.Msg)
	}

	fmt.Println("Selector changed successfully.")
	return nil
}
