package client

import (
	"errors"
	"fmt"
	"os"

	"github.com/jasonhancock/amproxy/pkg/amproxy"
	"github.com/jasonhancock/go-helpers"
	"github.com/spf13/cobra"
)

const envKeyPrivate = "KEY_PRIVATE"

func cmdSignature() *cobra.Command {
	var privateKey string

	cmd := &cobra.Command{
		Use:          "signature <message>",
		Short:        "Generates a signature for the provided input message",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if privateKey == "" {
				return errors.New("private key not specified")
			}

			msg, err := amproxy.Parse(args[0])
			if err != nil {
				return err
			}

			fmt.Println(msg.ComputeSignature(privateKey))

			return nil
		},
	}

	cmd.Flags().StringVar(
		&privateKey,
		"key-private",
		os.Getenv(envKeyPrivate),
		helpers.EnvDesc("The API private key.", envKeyPrivate),
	)

	return cmd
}
