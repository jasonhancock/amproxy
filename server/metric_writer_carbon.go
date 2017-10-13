package server

import (
	"net"

	"github.com/go-kit/kit/log"
	"github.com/jasonhancock/amproxy"
	"github.com/pkg/errors"
	"gopkg.in/fatih/pool.v2"
)

// MetricWriterCarbon will write metrics to a Carbon server. It pools connections.
type MetricWriterCarbon struct {
	logger log.Logger
	pool   pool.Pool
}

// NewMetricWriterCarbon creates a new MetricWriterCarbon
func NewMetricWriterCarbon(l log.Logger, addr string, poolMin, poolMax int) (*MetricWriterCarbon, error) {
	factory := func() (net.Conn, error) {
		l.Log("msg", "establishing_connection", "addr", addr)
		return net.Dial("tcp", addr)
	}

	pool, err := pool.NewChannelPool(poolMin, poolMax, factory)
	if err != nil {
		return nil, errors.Wrap(err, "creating pool")
	}

	m := &MetricWriterCarbon{
		logger: l,
		pool:   pool,
	}

	return m, nil
}

// WriteMetric writes the given message to the carbon server
func (mw *MetricWriterCarbon) WriteMetric(m amproxy.Message) error {
	conn, err := mw.pool.Get()
	if err != nil {
		return errors.Wrap(err, "getting connection from pool")
	}

	_, err = conn.Write([]byte(m.MetricStr() + "\n"))

	if err != nil {
		if pc, ok := conn.(*pool.PoolConn); ok {
			pc.MarkUnusable()
			pc.Close()
		}
		return errors.Wrap(err, "writing to connection")
	}

	conn.Close()
	return nil
}
