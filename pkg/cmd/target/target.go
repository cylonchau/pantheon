package target

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	targetExample = templates.Examples(i18n.T(`
		# Add a target
		pantheonctl target add --address 127.0.0.1:9090 --selector dc=ph190

		# To attach a label to the target.
		pantheonctl target add --address 127.0.0.1:9090 --labels=dc=prd-190 --selector dc=ph190

		# To attach multiple labels to the target.
		pantheonctl target add --address 127.0.0.1:9090 --labels dc=prd-190,module=a04-ws --selector dc=ph190

		# To attach multiple labels to the target.

		# List all targets.
		pantheonctl target list

		# List the targets for the specified instance (matched label).
		pantheonctl target list --labels dc=prd-190

		# Delete a target.
		pantheonctl target delete --id 1`))
)

type TargetLabel struct {
	Key   string
	Value string
}

type TargetAuth struct {
	Base        string `form:"base" json:"base" yaml:"base"`
	BearerToken string `form:"bearer_token" json:"bearer_token" yaml:"bearer_token"`
}

// NewCmdTarget creates a new Target command.
func NewCmdTarget() *cobra.Command {
	targetCmd := &cobra.Command{
		Use:                   "target",
		Short:                 "Manage targets",
		DisableFlagsInUseLine: true,
		Example:               targetExample,
	}
	targetAddCmd := newCmdTargetAdd()
	targetListCmd := newCmdTargetList()
	targetChangeCmd := newCmdTargetChange()
	targetDeleteCmd := newCmdTargetDelete()
	targetCleanCmd := newCmdTargetClean()
	targetAddFromFileCmd := newCmdTargetAddFromFile()
	targetCmd.AddCommand(
		targetAddCmd,
		targetListCmd,
		targetDeleteCmd,
		targetChangeCmd,
		targetAddFromFileCmd,
		targetCleanCmd,
	)
	return targetCmd
}

// validateKeyValuePairs ensures the key is valid and both key and value exist in the pair
func validateKeyValuePairs(pairs []TargetLabel, keyPattern string) error {

	for _, pair := range pairs {
		// Check if the key is empty
		if pair.Key == "" {
			return fmt.Errorf("label or selector key cannot be empty")
		}

		// Validate key format using regex pattern
		if match, _ := regexp.MatchString(keyPattern, pair.Key); !match {
			return fmt.Errorf("invalid key format: %s. Keys must start with a letter and can only contain letters, numbers, and hyphens (a-z,A-Z,0-9,_)", pair.Key)
		}

		// Ensure both key and value are provided (no standalone key allowed)
		if pair.Value == "" {
			return fmt.Errorf("value missing for key: %s. Both key and value must be provided in key=value format", pair.Key)
		}
	}
	return nil
}

func convertToRequestType(pairs []TargetLabel) map[string]string {
	resultMap := make(map[string]string)

	for _, s := range pairs {
		resultMap[s.Key] = s.Value
	}

	return resultMap
}

// confirmAndDelete wraps the deletion function with a confirmation prompt
func confirmAndExecute(yes bool, deleteFunc func() error, information string) error {
	if yes {
		return deleteFunc()
	}
	fmt.Println(information)
	fmt.Printf("Press Enter to continue or Ctrl+C to cancel:  Press Enter to continue or Ctrl+C to cancel: ")
	var input string
	fmt.Scanln(&input)

	if strings.TrimSpace(input) == "" {
		return deleteFunc()
	} else {
		fmt.Println("Deletion canceled.")
		return nil
	}
}
