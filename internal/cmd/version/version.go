package version

import "github.com/spf13/cobra"

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of this tool and exit",
		Args:  cobra.NoArgs,
		RunE:  run,
	}
}

func run(cmd *cobra.Command, args []string) error {
	cmd.Println("tool version: 1234")
	return nil
}
