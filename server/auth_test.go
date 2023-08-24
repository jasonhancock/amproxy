package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCredsAllowMetric(t *testing.T) {
	c := &Creds{
		Metrics: map[string]struct{}{
			"metric1": {},
			"metric2": {},
		},
	}

	require.True(t, c.AllowMetric("metric1"))
	require.True(t, c.AllowMetric("metric2"))
	require.False(t, c.AllowMetric("metric3"))
}
