package cmd

import (
	"context"
	"os"
	"sync"

	"github.com/jasonhancock/amproxy/cmd/client"
	"github.com/jasonhancock/amproxy/cmd/server"
	version "github.com/jasonhancock/cobra-version"
	"github.com/spf13/cobra"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context, wg *sync.WaitGroup, info version.Info) {
	rootCmd := newRootCmd(wg, info)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func newRootCmd(wg *sync.WaitGroup, info version.Info) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "amproxy",
		Short: "Authenticating Metrics Proxy",
	}

	cmd.AddCommand(
		client.NewCmd(),
		server.NewCmd(wg),
		version.NewCmd(info),
	)

	return cmd
}
