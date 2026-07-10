package monitor

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	monitorExample = templates.Examples(i18n.T(`
		# List all monitor rules.
		pantheonctl monitor list

		# Add a new monitor rule.
		pantheonctl monitor add --name=my-rule --type=pod --namespace=default --selector=app=test-app --port=metrics

		# Delete a monitor rule by ID.
		pantheonctl monitor delete 1`))
)

// NewCmdMonitor creates a new monitor command.
func NewCmdMonitor() *cobra.Command {
	monitorCmd := &cobra.Command{
		Use:                   "monitor",
		Short:                 "Manage monitor rules",
		DisableFlagsInUseLine: true,
		Example:               monitorExample,
	}

	monitorCmd.AddCommand(newCmdMonitorList())
	monitorCmd.AddCommand(newCmdMonitorAdd())
	monitorCmd.AddCommand(newCmdMonitorDelete())

	return monitorCmd
}
