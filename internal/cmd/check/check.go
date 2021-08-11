package check

import (
	"github.com/cdr/coder-doctor/internal/cmd/check/kubernetes"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "check that a resource is compatible with Coder",
		Args:  cobra.ExactArgs(1),
	}

	checkCmd.PersistentFlags().Int("verbosity", 0, "log level verbosity")
	checkCmd.PersistentFlags().String("coder-version", "1.21", "version of Coder")

	checkCmd.AddCommand(
		kubernetes.NewCommand(),
	)

	return checkCmd
}
