package cmd

import (
	"github.com/spf13/cobra"

	"github.com/cdr/coder-doctor/internal/cmd/check"
	"github.com/cdr/coder-doctor/internal/cmd/version"
)

func NewDefaultDoctorCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:  "coder-doctor",
		Args: cobra.ExactArgs(1),
	}
	rootCmd.AddCommand(
		version.NewCommand(),
		check.NewCommand(),
	)

	return rootCmd
}
