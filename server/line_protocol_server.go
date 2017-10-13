package server

import (
	"bufio"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

// LineProtocolServer is a generic server that takes input one line at a time
type LineProtocolServer struct {
	logger    log.Logger
	addr      string
	doneChan  chan struct{}
	wg        sync.WaitGroup
	listener  net.Listener
	processFn func(string)
}

// NewLineProtocolServer creates a new LineProtocolServer
func NewLineProtocolServer(l log.Logger, addr string, processFn func(string)) (*LineProtocolServer, error) {
	s := &LineProtocolServer{
		logger:    l,
		addr:      addr,
		doneChan:  make(chan struct{}),
		processFn: processFn,
	}

	return s, nil
}

// Run starts the server
func (s *LineProtocolServer) Run() error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return errors.Wrapf(err, "creating listener on %s", s.addr)
	}
	s.logger.Log("msg", "listening", "interface", s.addr)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.acceptConns(s.doneChan)
	}()

	return nil
}

// Stop shuts the server down, including all background go routines
func (s *LineProtocolServer) Stop() {
	s.logger.Log("msg", "stopping")
	close(s.doneChan)
	s.listener.Close()
	s.wg.Wait()
	s.logger.Log("msg", "stopped")
}

func (s *LineProtocolServer) acceptConns(done <-chan struct{}) error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-done:
				return nil
			default:
				s.logger.Log(
					"msg", "error_accepting_connection",
					"error", err,
				)
			}

			continue
		}
		// Handle connections in a new goroutine.
		go s.handleRequest(conn)
	}
}

func (s *LineProtocolServer) handleRequest(conn net.Conn) {
	defer conn.Close()
	s.logger.Log(
		"msg", "new_connection",
		"remote_ip", conn.RemoteAddr(),
	)

	r := bufio.NewReader(conn)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			s.logger.Log(
				"msg", "error_reading_conn",
				"error", err,
			)
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
