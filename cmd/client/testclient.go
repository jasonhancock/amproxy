package client

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/jasonhancock/amproxy/pkg/amproxy"
	"github.com/jasonhancock/amproxy/pkg/client"
	"github.com/jasonhancock/go-env"
	"github.com/jasonhancock/go-helpers"
	"github.com/spf13/cobra"
)

const envKeyPublic = "KEY_PUBLIC"

func cmdTestClient() *cobra.Command {
	var (
		privateKey string
		publicKey  string
		metricName string
		interval   time.Duration
		serverAddr string
	)

	cmd := &cobra.Command{
		Use:          "test-client <message>",
		Short:        "sends messages for a metric for testing",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if privateKey == "" {
				return errors.New("private key not specified")
			}

			if publicKey == "" {
				return errors.New("public key not specified")
			}

			ticker := time.NewTimer(0)
			min := 30
			max := 100

			cl := client.NewClient(publicKey, privateKey, serverAddr)

			for {
				select {
				case <-cmd.Context().Done():
					ticker.Stop()
					return nil
				case <-ticker.C:
					log.Println("sending")
					m := amproxy.Message{
						Name:      metricName,
						Value:     fmt.Sprintf("%d", rand.Intn(max-min)+min),
						Timestamp: time.Now(),
					}

					if err := cl.Connect(); err != nil {
						return err
					}

					if err := cl.Write(m); err != nil {
						return err
					}

					if err := cl.Disconnect(); err != nil {
						return err
					}

					ticker.Reset(interval)
				}
			}
		},
	}

	cmd.Flags().StringVar(
		&privateKey,
		"key-private",
		os.Getenv(envKeyPrivate),
		helpers.EnvDesc("The API private key.", envKeyPrivate),
	)

	cmd.Flags().StringVar(
		&publicKey,
		"key-public",
		os.Getenv(envKeyPublic),
		helpers.EnvDesc("The API public key.", envKeyPublic),
	)

	const envMetricName = "METRIC_NAME"
	cmd.Flags().StringVar(
		&metricName,
		"metric",
		env.String(envMetricName, "metric1"),
		helpers.EnvDesc("The name of the metric to send.", envMetricName),
	)

	const envServerAddr = "SERVER_ADDR"
	cmd.Flags().StringVar(
		&serverAddr,
		"server-addr",
		env.String(envServerAddr, "127.0.0.1:2005"),
		helpers.EnvDesc("The server address to send to.", envServerAddr),
	)

	const envInterval = "INTERVAL"
	cmd.Flags().DurationVar(
		&interval,
		"interval",
		env.Duration(envInterval, 1*time.Minute),
		helpers.EnvDesc("How often to send a message", envInterval),
	)

	return cmd
}
