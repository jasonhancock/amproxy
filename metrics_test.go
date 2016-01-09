package amproxy_test

import (
	"testing"

	"github.com/jasonhancock/amproxy"
)

func TestReverseString(t *testing.T) {
	if amproxy.ReverseDelimitedString("foo.bar.baz", ".") != "baz.bar.foo" {
		t.Errorf("reverse delimited string failed")
	}
}
