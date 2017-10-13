package server

import (
	"os"
	"testing"
	"time"

	"github.com/cheekybits/is"
	"github.com/go-kit/kit/log"
	"github.com/jasonhancock/amproxy"
)

func TestServer(t *testing.T) {
	is := is.New(t)

	var observedMetrics []string

	logger := testLogger()
	mw := &mockMetricWriter{
		WriteMetricFn: func(m amproxy.Message) error {
			observedMetrics = append(observedMetrics, m.Name)
			return nil
		},
	}

	publicKey := "public_key"
	privateKey := "private_key"
	metricName := "foo.bar"

	creds := map[string]*Creds{
		publicKey: &Creds{
			SecretKey: privateKey,
			Metrics: map[string]uint8{
				metricName: 1,
			},
		},
	}

	ap := &mockAuthProvider{
		CredsForKeyFn: func(key string) (*Creds, error) {
			c, ok := creds[key]
			if !ok {
				return nil, ErrCredentialsNotFound
			}
			return c, nil
		},
	}

	s, err := NewServer(log.With(logger, "component", "server"), ":8095", 300, ap, mw)
	is.NoErr(err)
	is.NoErr(s.Run())
	defer s.Stop()
	time.Sleep(300 * time.Millisecond)

	client := amproxy.NewClient(publicKey, privateKey, ":8095")
	m := amproxy.Message{
		Name:      metricName,
		Value:     "1234",
		Timestamp: int(time.Now().Unix()),
	}
	client.Write(m)
	m.Name = "foo.bar.baz"
	client.Write(m)
	time.Sleep(300 * time.Millisecond)
}

func testLogger() log.Logger {
	return log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
}
