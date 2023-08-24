package server

import (
	"errors"
	"testing"
	"time"

	"github.com/jasonhancock/amproxy/pkg/amproxy"
	"github.com/jasonhancock/go-logger"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	const publicKey = "public_key"
	const privateKey = "private_key"
	const metricName = "foo.bar"

	tests := []struct {
		description     string
		inputMetricName string
		inputSkew       time.Duration
		inputPubKey     string
		inputPrivateKey string
		errMetricWriter error
		err             error
		expected        []amproxy.Message
	}{
		{
			"normal",
			metricName,
			0 * time.Second,
			publicKey,
			privateKey,
			nil,
			nil,
			[]amproxy.Message{
				{
					Name:  metricName,
					Value: "1234",
				},
			},
		},
		{
			"metric writer error",
			metricName,
			0 * time.Second,
			publicKey,
			privateKey,
			errors.New("some mw error"),
			errors.New("write metric: some mw error"),
			nil,
		},
		{
			"bad line",
			"",
			0 * time.Second,
			publicKey,
			privateKey,
			nil,
			errors.New("decomposing message"),
			nil,
		},
		{
			"bad metric name",
			metricName + "foo",
			0 * time.Second,
			publicKey,
			privateKey,
			nil,
			errors.New("not an approved metric"),
			nil,
		},
		{
			"bad clock skew",
			metricName,
			600 * time.Second,
			publicKey,
			privateKey,
			nil,
			errors.New("max skew set to 5m0s"),
			nil,
		},
		{
			"bad pub key",
			metricName,
			0 * time.Second,
			publicKey + "bad",
			privateKey,
			nil,
			ErrCredentialsNotFound,
			nil,
		},
		{
			"bad signature",
			metricName,
			0 * time.Second,
			publicKey,
			privateKey + "bad",
			nil,
			errors.New("doesn't match provided signature"),
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			var observedMetrics []amproxy.Message
			mw := &mockMetricWriter{
				WriteMetricFn: func(m amproxy.Message) error {
					observedMetrics = append(observedMetrics, m)
					return tt.errMetricWriter
				},
			}

			creds := map[string]*Creds{
				publicKey: {
					SecretKey: privateKey,
					Metrics: map[string]struct{}{
						metricName: {},
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

			l := logger.Default()

			s := NewServer(l, 300*time.Second, ap, mw)

			m := amproxy.Message{
				Name:      tt.inputMetricName,
				Value:     "1234",
				Timestamp: time.Now().Add(tt.inputSkew),
				PublicKey: tt.inputPubKey,
			}
			m.Signature = m.ComputeSignature(tt.inputPrivateKey)

			err := s.processLine(m.String())
			if tt.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.err.Error())
				return
			}
			require.NoError(t, err)

			require.Len(t, observedMetrics, len(tt.expected))
			for i := range tt.expected {
				require.Equal(t, tt.expected[i].Name, observedMetrics[0].Name)
				require.Equal(t, tt.expected[i].Value, observedMetrics[0].Value)
			}
		})
	}
}
