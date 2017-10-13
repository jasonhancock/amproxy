package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jasonhancock/amproxy/server"
)

func main() {
	var (
		logger log.Logger

		addr       = flag.String("addr", ":2005", "interface/port to bind to")
		carbonAddr = flag.String("carbon-addr", "127.0.0.1:2003", "Carbon address:port")
		authFile   = flag.String("auth-file", "/etc/amproxy/auth_file.yaml", "Location of auth file")
		skew       = flag.Float64("skew", 300, "amount of clock skew tolerated in seconds")
	)
	flag.Parse()

	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	if *authFile == "" {
		logger.Log("msg", "missing_auth_file")
		os.Exit(1)
	}

	ap, err := server.NewAuthProviderStaticFile(log.With(logger, "component", "auth_provider"), *authFile, 1*time.Minute)
	if err != nil {
		logger.Log("msg", "constructing_auth_provider_error", "error", err)
		os.Exit(1)
	}

	mw, err := server.NewMetricWriterCarbon(log.With(logger, "component", "metric_writer"), *carbonAddr, 0, 30)
	if err != nil {
		logger.Log("msg", "constructing_metric_writer_error", "error", err)
		os.Exit(1)
	}

	s, err := server.NewServer(log.With(logger, "component", "server"), *addr, *skew, ap, mw)
	if err != nil {
		logger.Log("msg", "constructing_server_error", "error", err)
		os.Exit(1)
	}

	// subscribe to signals to shut down the server
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)
	signal.Notify(stopChan, os.Kill)
	signal.Notify(stopChan, syscall.SIGTERM)

	err = s.Run()
	if err != nil {
		logger.Log("msg", "running_server_error", "error", err)
		os.Exit(1)
	}
	<-stopChan // wait for signals
	s.Stop()
}
