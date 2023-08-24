package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jasonhancock/go-logger"
	"github.com/stretchr/testify/require"
)

func TestAuthProviderStaticFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "auth.yaml")
	require.NoError(t, os.WriteFile(file, []byte(authData), 0644))

	interval := 100 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a, err := NewAuthProviderStaticFile(ctx, logger.Silence(), file, interval)
	require.NoError(t, err)

	c, err := a.CredsForKey("apikey")
	require.NoError(t, err)
	require.True(t, c.AllowMetric("metric1"))
	require.False(t, c.AllowMetric("metric3"))

	_, err = a.CredsForKey("apikey2")
	require.Equal(t, err, ErrCredentialsNotFound)

	time.Sleep(interval)
	require.NoError(t, os.WriteFile(file, []byte(authData2), 0644))
	time.Sleep(2 * interval)

	c, err = a.CredsForKey("apikey")
	require.NoError(t, err)
	require.True(t, c.AllowMetric("metric1"))
	require.True(t, c.AllowMetric("metric3"))
}

const authData = `
---
apikeys:
  apikey:
    secret_key: blah
    metrics:
    - metric1
    - metric2
`

const authData2 = `
---
apikeys:
  apikey:
    secret_key: blah
    metrics:
    - metric1
    - metric2
    - metric3
`
