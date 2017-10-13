package server

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cheekybits/is"
)

func TestAuthProviderStaticFile(t *testing.T) {
	is := is.New(t)
	logger := testLogger()

	dir, err := ioutil.TempDir("", "")
	is.NoErr(err)
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "auth.yaml")
	is.NoErr(ioutil.WriteFile(file, []byte(authData), 0644))

	done := make(chan struct{})

	interval := 1 * time.Second

	a, err := NewAuthProviderStaticFile(logger, file, interval)
	is.NoErr(err)
	go a.Run(done)
	defer close(done)

	c, err := a.CredsForKey("apikey")
	is.NoErr(err)
	is.True(c.AllowMetric("metric1"))
	is.False(c.AllowMetric("metric3"))

	_, err = a.CredsForKey("apikey2")
	is.Equal(err, ErrCredentialsNotFound)

	time.Sleep(1 * time.Second)
	is.NoErr(ioutil.WriteFile(file, []byte(authData2), 0644))
	time.Sleep(2 * interval)

	c, err = a.CredsForKey("apikey")
	is.NoErr(err)
	is.True(c.AllowMetric("metric1"))
	is.True(c.AllowMetric("metric3"))
}

const authData = `
---
apikeys:
  apikey:
    secret_key: blah
    metrics:
      metric1: 1
      metric2: 1
`

const authData2 = `
---
apikeys:
  apikey:
    secret_key: blah
    metrics:
      metric1: 1
      metric2: 1
      metric3: 1
`
