package amproxy

import (
	"testing"

	"github.com/cheekybits/is"
)

func TestReverseString(t *testing.T) {
	is := is.New(t)
	is.Equal(reverseDelimitedString("foo.bar.baz", "."), "baz.bar.foo")
}
