package server

import (
	"testing"
	"time"

	"github.com/cheekybits/is"
	"github.com/go-kit/kit/log"
	"github.com/jasonhancock/amproxy"
)

func TestMetricWriterCarbon(t *testing.T) {
	is := is.New(t)
	logger := testLogger()

	var lines []string

	fn := func(line string) {
		lines = append(lines, line)
	}

	s, err := NewLineProtocolServer(log.With(logger, "component", "line_protocol_server"), ":8096", fn)
	is.NoErr(err)
	is.NoErr(s.Run())
	defer s.Stop()

	mw, err := NewMetricWriterCarbon(logger, ":8096", 0, 3)
	is.NoErr(err)

	m := amproxy.Message{
		Name:      "test.metric.foo",
		Value:     "1234",
		Timestamp: 3456,
	}

	mw.WriteMetric(m)
	m.Value = "4321"
	m.Timestamp = 3457
	mw.WriteMetric(m)
	time.Sleep(300 * time.Millisecond)

	is.Equal(len(lines), 2)
	is.Equal(lines[0], "test.metric.foo 1234 3456")
	is.Equal(lines[1], "test.metric.foo 4321 3457")
}
