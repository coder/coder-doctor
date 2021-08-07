package check

import (
	"github.com/cdr/doctor/internal/cmd/check/kubernetes"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "scan the Kubernetes cluster for compatibility",
		Args:  cobra.ExactArgs(1),
	}

	checkCmd.PersistentFlags().Int("verbosity", 0, "log level verbosity")

	checkCmd.AddCommand(
		kubernetes.NewCommand(),
	)

	return checkCmd
}
