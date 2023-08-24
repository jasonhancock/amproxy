package amproxy

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseSigned(t *testing.T) {

	t.Run("invalid number of components", func(t *testing.T) {
		_, err := ParseSigned("foo 1234 1425059762 my_public_key")
		require.Equal(t, ErrInvalidNumMessageComponents, err)
	})

	t.Run("timestamp not a number", func(t *testing.T) {
		_, err := ParseSigned("foo 1234 abc1425059762 my_public_key lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=")
		require.Error(t, err)
		_, ok := err.(*strconv.NumError)
		require.True(t, ok)
	})

	t.Run("normal", func(t *testing.T) {
		m, err := ParseSigned("foo 1234 1425059762 my_public_key lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=")
		require.NoError(t, err)
		require.Equal(t, "foo", m.Name)
		require.Equal(t, "1234", m.Value)
		require.Equal(t, time.Unix(1425059762, 0), m.Timestamp)
		require.Equal(t, "my_public_key", m.PublicKey)
		require.Equal(t, "lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=", m.Signature)
	})
}

func TestComputeSignature(t *testing.T) {
	m, err := ParseSigned("foo 1234 1425059762 my_public_key lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=")
	require.NoError(t, err)
	require.Equal(t, "lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=", m.ComputeSignature("my_secret_key"))
}

func TestStrings(t *testing.T) {
	m, err := ParseSigned("foo 1234 1425059762 my_public_key lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=")
	require.NoError(t, err)

	require.Equal(t, "foo 1234 1425059762", m.MetricStr())
	require.Equal(t, "foo 1234 1425059762 my_public_key lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=", m.String())
}
