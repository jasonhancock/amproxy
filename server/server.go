package server

import (
	"math"
	"net"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jasonhancock/amproxy"
	"github.com/pkg/errors"
)

// Server is an amproxy server. It uses a generic LineProtocolServer to handle
// the connections, but implements the business logic itself.
type Server struct {
	logger       log.Logger
	authProvider AuthProvider
	skew         float64
	doneChan     chan struct{}
	wg           sync.WaitGroup
	listener     net.Listener
	metricWriter MetricWriter
	server       *LineProtocolServer
}

// NewServer creates a new Server
func NewServer(l log.Logger, addr string, skew float64, authProvider AuthProvider, mw MetricWriter) (*Server, error) {
	s := &Server{
		logger:       l,
		skew:         skew,
		authProvider: authProvider,
		doneChan:     make(chan struct{}),
		metricWriter: mw,
	}

	lp, err := NewLineProtocolServer(log.With(l, "component", "line_protocol_server"), addr, func(line string) {
		if err := s.processLine(line); err != nil {
			l.Log("msg", "process_line_error", "error", err)
		}
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating protocol server")
	}
	s.server = lp

	return s, nil
}

// Run starts the Server
func (s *Server) Run() error {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.authProvider.Run(s.doneChan)
	}()

	err := s.server.Run()
	if err != nil {
		return errors.Wrapf(err, "running lp server")
	}

	return nil
}

// Stop shuts the server down, including all background go routines
func (s *Server) Stop() {
	s.logger.Log("msg", "stopping")
	close(s.doneChan)
	s.server.Stop()
	s.wg.Wait()
	s.logger.Log("msg", "stopped")
}

func (s *Server) processLine(line string) error {

	msg, err := amproxy.Parse(line)
	if err != nil {
		return errors.Wrapf(err, "decomposing message %q", line)
	}

	creds, err := s.authProvider.CredsForKey(msg.PublicKey)
	if err != nil {
		return errors.Wrapf(err, "pub key %q", msg.PublicKey)
	}

	sig := msg.ComputeSignature(creds.SecretKey)

	if sig != msg.Signature {
		return errors.Errorf("computed signature %s doesn't match provided signature %s", sig, msg.Signature)
	}

	delta := math.Abs(float64(time.Now().Unix() - int64(msg.Timestamp)))
	if delta > s.skew {
		return errors.Errorf("delta = %.0f, max skew set to %.0f", delta, s.skew)
	}

	// validate the metric is on the approved list
	if !creds.AllowMetric(msg.Name) {
		return errors.Errorf("not an approved metric: %s", msg.Name)
	}

	if err := s.metricWriter.WriteMetric(*msg); err != nil {
		return errors.Wrap(err, "write metric")
	}
	return nil
}
