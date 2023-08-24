package lineserver

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/jasonhancock/go-logger"
)

// LineProtocolServer is a generic server that takes input one line at a time.
type LineProtocolServer struct {
	logger    *logger.L
	listener  net.Listener
	processFn func(string)
}

// New creates a new LineProtocolServer
func New(ctx context.Context, l *logger.L, wg *sync.WaitGroup, addr string, processFn func(string)) (*LineProtocolServer, error) {
	s := &LineProtocolServer{
		logger:    l,
		processFn: processFn,
	}

	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("creating listener on %q: %w", addr, err)
	}
	s.logger.Info("listening", "addr", addr)

	go func() {
		<-ctx.Done()
		s.logger.Info("closing listener")
		s.listener.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.acceptConns(ctx)
	}()

	return s, nil
}

func (s *LineProtocolServer) acceptConns(ctx context.Context) error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				s.logger.LogError("error_accepting_connection", err)
			}

			continue
		}
		// Handle connections in a new goroutine.
		go s.handleRequest(ctx, conn)
	}
}

func (s *LineProtocolServer) handleRequest(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	s.logger.Info(
		"new_connection",
		"remote_ip", conn.RemoteAddr(),
	)

	r := bufio.NewReader(conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			s.logger.LogError("error_reading_conn", err)
			return
		}
		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		if s.processFn != nil {
			s.processFn(line)
		}
	}
}
