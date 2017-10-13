package server

import (
	"testing"

	"github.com/cheekybits/is"
)

func TestCredsAllowMetric(t *testing.T) {
	is := is.New(t)

	c := &Creds{
		Metrics: map[string]uint8{
			"metric1": 1,
			"metric2": 1,
		},
	}

	is.True(c.AllowMetric("metric1"))
	is.True(c.AllowMetric("metric2"))
	is.False(c.AllowMetric("metric3"))
}
