package selector

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/cylonchau/pantheon/pkg/api/selector"
)

var (
	selectorExample = templates.Examples(i18n.T(`
		# List all selectors.
		pantheonctl selector list`))
)

// NewCmdselector creates a new selector command.
func NewCmdselector() *cobra.Command {
	selectorCmd := &cobra.Command{
		Use:                   "selector",
		Short:                 "Manage selectors",
		DisableFlagsInUseLine: true,
		Example:               selectorExample,
	}
	selectorListCmd := newCmdselectorList()
	selectorChgCmd := newCmdselectorChange()
	selectorCmd.AddCommand(selectorListCmd, selectorChgCmd)
	return selectorCmd
}

// validateKeyValuePairs ensures the key is valid and both key and value exist in the pair
func validateKeyValuePairs(pairs []selector.SelectorItem, keyPattern string) error {

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

func convertToRequestType(pairs []selector.SelectorItem) map[string]string {
	resultMap := make(map[string]string)

	for _, s := range pairs {
		resultMap[s.Key] = s.Value
	}

	return resultMap
}
