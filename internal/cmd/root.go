package cmd

import (
	"github.com/spf13/cobra"

	"cdr.dev/coder-doctor/internal/cmd/check"
	"cdr.dev/coder-doctor/internal/cmd/version"
)

func NewDefaultDoctorCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "coder-doctor",
		Short: "coder-doctor checks compatibility with Coder",
		Long:  `coder-doctor is a tool for analyzing that Coder's dependencies satisfy our requirements.`,
		Args:  cobra.ExactArgs(1),
	}
	rootCmd.AddCommand(
		version.NewCommand(),
		check.NewCommand(),
	)

	rootCmd.PersistentFlags().Bool("output-colors", true, "enable colorful output")
	rootCmd.PersistentFlags().Bool("output-ascii", false, "output ascii only")

	return rootCmd
}
