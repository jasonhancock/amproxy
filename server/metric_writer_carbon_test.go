package server

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jasonhancock/amproxy/pkg/amproxy"
	"github.com/jasonhancock/amproxy/pkg/lineserver"
	"github.com/jasonhancock/go-logger"
	"github.com/jasonhancock/go-testhelpers/generic"
	"github.com/stretchr/testify/require"
)

func TestMetricWriterCarbon(t *testing.T) {
	l := logger.Silence()

	var lines []string
	fn := func(line string) {
		lines = append(lines, line)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	addr := generic.NewRandomPort(t)
	_, err := lineserver.New(ctx, l, &wg, addr, fn)
	require.NoError(t, err)

	mw, err := NewMetricWriterCarbon(l, addr, 0, 3)
	require.NoError(t, err)

	m := amproxy.Message{
		Name:      "test.metric.foo",
		Value:     "1234",
		Timestamp: time.Unix(3456, 0),
	}

	mw.WriteMetric(m)
	m.Value = "4321"
	m.Timestamp = time.Unix(3457, 0)
	mw.WriteMetric(m)

	time.Sleep(300 * time.Millisecond)

	require.Len(t, lines, 2)
	require.Equal(t, "test.metric.foo 1234 3456", lines[0])
	require.Equal(t, "test.metric.foo 4321 3457", lines[1])
}
