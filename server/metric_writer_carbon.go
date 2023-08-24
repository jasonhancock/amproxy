package server

import (
	"fmt"
	"net"

	"github.com/jasonhancock/amproxy/pkg/amproxy"
	"github.com/jasonhancock/go-logger"
	"gopkg.in/fatih/pool.v2"
)

// MetricWriterCarbon will write metrics to a Carbon server. It pools connections.
type MetricWriterCarbon struct {
	pool pool.Pool
}

// NewMetricWriterCarbon creates a new MetricWriterCarbon
func NewMetricWriterCarbon(l *logger.L, addr string, poolMin, poolMax int) (*MetricWriterCarbon, error) {
	factory := func() (net.Conn, error) {
		l.Info("establishing_connection", "addr", addr)
		return net.Dial("tcp", addr)
	}

	pool, err := pool.NewChannelPool(poolMin, poolMax, factory)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	m := &MetricWriterCarbon{
		pool: pool,
	}

	return m, nil
}

// WriteMetric writes the given message to the carbon server
func (mw *MetricWriterCarbon) WriteMetric(m amproxy.Message) error {
	conn, err := mw.pool.Get()
	if err != nil {
		return fmt.Errorf("getting connection from pool: %w", err)
	}

	_, err = conn.Write([]byte(m.MetricStr() + "\n"))

	if err != nil {
		if pc, ok := conn.(*pool.PoolConn); ok {
			pc.MarkUnusable()
			pc.Close()
		}
		return fmt.Errorf("writing to connection: %w", err)
	}

	return conn.Close()
}
