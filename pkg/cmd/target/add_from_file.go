package target

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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

var (
	addFromFileExample = templates.Examples(i18n.T(`
		# Delete a target
		pantheonctl target add-from-file -f xxx.yaml`))
)

// TargetDeleteOptions holds the options for the delete command
type TargetAddFromFileOptions struct {
	filePath string
}

// NewTargetDeleteOptions creates the options for the delete command
func NewTargetAddFromFileOptions() *TargetAddFromFileOptions {
	return &TargetAddFromFileOptions{}
}

// NewCmdTargetAddFromFile creates a new command to add targets from a file
func newCmdTargetAddFromFile() *cobra.Command {
	o := NewTargetAddFromFileOptions()
	cmd := &cobra.Command{
		Use:     "add-from-file",
		Short:   "Add targets from a YAML/JSON file",
		Aliases: []string{"f", "file"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	cmd.Flags().StringVarP(&o.filePath, "file", "f", "", "Path to the file containing targets to add (YAML/JSON)")
	cmd.MarkFlagRequired("file")

	return cmd
}

func (o *TargetAddFromFileOptions) Run() error {
	// 读取文件内容
	data, err := os.ReadFile(o.filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %s", err)
	}

	// 文件格式为 YAML，解析成目标列表
	var targets target.Target
	if err := yaml.Unmarshal(data, &targets); err != nil {
		return fmt.Errorf("failed to parse YAML: %s", err)
	}

	if err := o.addSingleTarget(targets); err != nil {
		return err
	}
	return nil
}

func (o *TargetAddFromFileOptions) addSingleTarget(target target.Target) error {
	cluster, err := config.GetClusterConfig()

	body, err := sonic.Marshal(target)
	if err != nil {
		return err
	}

	api, exists := path_map.APIInterfaces["AddTarget"]
	if !exists {
		return fmt.Errorf("Unsupport API %s", api.Path)
	}
	url := fmt.Sprintf("%s%s", cluster.Cluster.Server, api.Path)
	// Send the HTTP request
	resp, err := utils.SendRequest(api.Method, url, body, cluster.Cluster.Auth) // Directly pass the Auth info
	if err != nil {
		return err
	}
	defer resp.Body.Close()
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

		return fmt.Errorf("failed to process bulk targets: %s", responseBody.Msg)
	}

	fmt.Printf("bulk targets created\n")
	return nil
}
