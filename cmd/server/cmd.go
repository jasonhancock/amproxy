package server

import (
	"os"
	"sync"
	"time"

	"github.com/jasonhancock/amproxy/pkg/lineserver"
	"github.com/jasonhancock/amproxy/server"
	clog "github.com/jasonhancock/cobra-logger"
	"github.com/jasonhancock/go-env"
	"github.com/jasonhancock/go-helpers"
	"github.com/spf13/cobra"
)

const defaultAuthFile = "/etc/amproxy/auth_file.yaml"

func NewCmd(wg *sync.WaitGroup) *cobra.Command {
	var (
		addr       string
		carbonAddr string
		authFile   string
		skew       time.Duration
		logConf    *clog.Config
	)

	cmd := &cobra.Command{
		Use:          "server",
		Short:        "Starts the server",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			l := logConf.Logger(os.Stdout)

			if authFile == "" {
				l.Fatal("missing auth file")
			}

			ap, err := server.NewAuthProviderStaticFile(cmd.Context(), l.New("auth_provider"), authFile, 1*time.Minute)
			if err != nil {
				l.Fatal("constructing_auth_provider_error", "error", err)
			}

			mw, err := server.NewMetricWriterCarbon(l.New("metric_writer"), carbonAddr, 0, 30)
			if err != nil {
				l.Fatal("constructing_metric_writer_error", "error", err)
			}

			s := server.NewServer(l.New("server"), skew, ap, mw)

			_, err = lineserver.New(cmd.Context(), l.New("line_protocol_server"), wg, addr, s.ProcessLine)
			if err != nil {
				l.Fatal("starting line protocol server error", "error", err)
			}

			wg.Wait()
			return nil
		},
	}

	logConf = clog.NewConfig(cmd)

	const envAddr = "ADDR"
	cmd.Flags().StringVar(
		&addr,
		"addr",
		env.String(envAddr, "127.0.0.1:2005"),
		helpers.EnvDesc("The interface and port to bind the server to.", envAddr),
	)

	const envCarbonAddr = "CARBON_ADDR"
	cmd.Flags().StringVar(
		&carbonAddr,
		"carbon-addr",
		env.String(envCarbonAddr, "127.0.0.1:2003"),
		helpers.EnvDesc("The address of the carbon server.", envCarbonAddr),
	)

	const envAuthFile = "AUTH_FILE"
	cmd.Flags().StringVar(
		&authFile,
		"auth-file",
		env.String(envAuthFile, defaultAuthFile),
		helpers.EnvDesc("The path to the auth file.", envAuthFile),
	)

	const envSkew = "MAX_SKEW"
	cmd.Flags().DurationVar(
		&skew,
		"skew",
		env.Duration(envSkew, 300*time.Second),
		helpers.EnvDesc("The amount of clock skew tolerated.", envSkew),
	)

	return cmd
}
