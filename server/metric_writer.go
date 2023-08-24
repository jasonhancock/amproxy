package server

import "github.com/jasonhancock/amproxy/pkg/amproxy"

// MetricWriter is the interface for a backend metrics store
type MetricWriter interface {
	WriteMetric(amproxy.Message) error
}

type mockMetricWriter struct {
	WriteMetricFn func(amproxy.Message) error
}

func (mw *mockMetricWriter) WriteMetric(m amproxy.Message) error {
	if mw.WriteMetricFn != nil {
		return mw.WriteMetricFn(m)
	}
	panic("not implemented")
}
