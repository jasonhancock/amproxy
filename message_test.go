package amproxy

import (
	"testing"

	"github.com/cheekybits/is"
)

func TestDecompose(t *testing.T) {
	is := is.New(t)

	m := &Message{}
	m, err := Decompose("foo 1234 1425059762 my_public_key")
	is.Err(err)
	is.Equal(err, errInvalidNumMessageComponents)

	m, err = Decompose("foo 1234 1425059762 my_public_key lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=")
	is.NoErr(err)
	is.Equal(m.Name, "foo")
	is.Equal(m.Value, "1234")
	is.Equal(m.Timestamp, 1425059762)
	is.Equal(m.PublicKey, "my_public_key")
	is.Equal(m.Signature, "lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=")
}

func TestComputeSignature(t *testing.T) {
	is := is.New(t)

	m, err := Decompose("foo 1234 1425059762 my_public_key lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=")
	is.NoErr(err)
	is.Equal(m.ComputeSignature("my_secret_key"), "lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=")
}

func TestMetricStr(t *testing.T) {
	is := is.New(t)

	m, err := Decompose("foo 1234 1425059762 my_public_key lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=")
	is.NoErr(err)
	is.Equal(m.MetricStr(), "foo 1234 1425059762")
}
