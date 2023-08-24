package client

import (
	"github.com/spf13/cobra"
)

// NewCmd initializes a new command and sub-commands.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Client operations.",
	}

	cmd.AddCommand(
		cmdSignature(),
		cmdTestClient(),
	)

	return cmd
}
