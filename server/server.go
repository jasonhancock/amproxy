package server

import (
	"fmt"
	"math"
	"time"

	"github.com/jasonhancock/amproxy/pkg/amproxy"
	"github.com/jasonhancock/go-logger"
)

// Server is an amproxy server. It uses a generic LineProtocolServer to handle
// the connections, but implements the business logic itself.
type Server struct {
	logger       *logger.L
	authProvider AuthProvider
	skew         time.Duration
	metricWriter MetricWriter
}

// NewServer creates a new Server
func NewServer(l *logger.L, skew time.Duration, authProvider AuthProvider, mw MetricWriter) *Server {
	s := &Server{
		logger:       l,
		skew:         skew,
		authProvider: authProvider,
		metricWriter: mw,
	}

	return s
}

func (s *Server) ProcessLine(line string) {
	if err := s.processLine(line); err != nil {
		s.logger.LogError("process_line_error", err)
	}
}

func (s *Server) processLine(line string) error {
	msg, err := amproxy.ParseSigned(line)
	if err != nil {
		return fmt.Errorf("decomposing message %q: %w", line, err)
	}

	creds, err := s.authProvider.CredsForKey(msg.PublicKey)
	if err != nil {
		return fmt.Errorf("pub key %q: %w", msg.PublicKey, err)
	}

	sig := msg.ComputeSignature(creds.SecretKey)
	if sig != msg.Signature {
		return fmt.Errorf("computed signature %s doesn't match provided signature %s", sig, msg.Signature)
	}

	delta := time.Duration(math.Abs(float64(time.Since(msg.Timestamp))))
	if delta > s.skew {
		return fmt.Errorf("delta = %s, max skew set to %s", delta, s.skew)
	}

	// validate the metric is on the approved list
	if !creds.AllowMetric(msg.Name) {
		return fmt.Errorf("not an approved metric: %s", msg.Name)
	}

	if err := s.metricWriter.WriteMetric(*msg); err != nil {
		return fmt.Errorf("write metric: %w", err)
	}
	return nil
}
