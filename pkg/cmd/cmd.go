package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/cylonchau/pantheon/pkg/cmd/config"
	"github.com/cylonchau/pantheon/pkg/cmd/selector"
	"github.com/cylonchau/pantheon/pkg/cmd/target"
)

type PantheonctlOptions struct {
	Arguments []string
}

// NewDefaultKubectlCommand creates the `kubectl` command with default arguments
func NewDefaultPantheonctlCommand() *cobra.Command {
	return NewDefaultPantheonctlCommandWithArgs(PantheonctlOptions{
		Arguments: os.Args,
	})
}

// NewDefaultKubectlCommandWithArgs creates the `kubectl` command with arguments
func NewDefaultPantheonctlCommandWithArgs(o PantheonctlOptions) *cobra.Command {
	cmd := NewPantheonctlCommand(o)

	if len(o.Arguments) > 1 {
		cmdPathPieces := o.Arguments[1:]

		// only look for suitable extension executables if
		// the specified command does not already exist
		if _, _, err := cmd.Find(cmdPathPieces); err != nil {
			// Also check the commands that will be added by Cobra.
			// These commands are only added once rootCmd.Execute() is called, so we
			// need to check them explicitly here.
			var cmdName string // first "non-flag" arguments
			for _, arg := range cmdPathPieces {
				if !strings.HasPrefix(arg, "-") {
					cmdName = arg
					break
				}
			}

			switch cmdName {
			case "help", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
				// Don't search for a plugin
			default:

			}
		}
	}

	return cmd
}

// NewKubectlCommand creates the `kubectl` command and its nested children.
func NewPantheonctlCommand(o PantheonctlOptions) *cobra.Command {
	// Parent command to which all subcommands are added.
	rootCmd := &cobra.Command{
		Use:   "pantheonctl",
		Short: i18n.T("pantheonctl controls the Pantheon cluster manager"),
		Long: templates.Examples(i18n.T(`
      Pantheon controls the Pantheon cluster manager.

      Find more information at:
            https://github.com/cylonchau/pantheon`)),
		Run: runHelp,
	}

	targetCmd := target.NewCmdTarget()
	configCmd := config.NewCmdConfig()
	selectorCmd := selector.NewCmdselector()
	rootCmd.AddCommand(
		targetCmd,
		configCmd,
		selectorCmd,
	)
	return rootCmd
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}
