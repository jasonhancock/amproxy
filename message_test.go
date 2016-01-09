package amproxy_test

import (
	"testing"

	"github.com/jasonhancock/amproxy"
)

func TestDecompose(t *testing.T) {

	m := new(amproxy.Message)
	err := m.Decompose("foo 1234 1425059762 my_public_key")
	if err == nil {
		t.Errorf("invalid num of components. Expected error, didn't get it")
	}

	err2 := m.Decompose("foo 1234 1425059762 my_public_key lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=")

	if err2 != nil {
		t.Errorf("Didn't expect an error")
	}

	if m.Name != "foo" {
		t.Errorf("Name expected: %q got: %q", "foo", m.Name)
	}

	if m.Value != "1234" {
		t.Errorf("Value expected: %s got: %s", "1234", m.Value)
	}

	if m.Timestamp != 1425059762 {
		t.Errorf("Timestamp expected: %d got: %d", 1425059762, m.Timestamp)
	}

	if m.Public_key != "my_public_key" {
		t.Errorf("Public_key expected: %q got: %q", "my_public_key", m.Public_key)
	}

	if m.Signature != "lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=" {
		t.Errorf("Signature expected: %q got: %q", "lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=", m.Signature)
	}
}

func TestComputeSignature(t *testing.T) {
	var expected string = "lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc="
	m := new(amproxy.Message)
	err := m.Decompose("foo 1234 1425059762 my_public_key lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=")
	if err != nil {
		t.Errorf("Shouldn't have gotten an error")
	}
	got := m.ComputeSignature("my_secret_key")

	if got != expected {
		t.Errorf("ComputeSignature expected: %q got: %q", expected, got)
	}
}

func TestMetricStr(t *testing.T) {
	var expected string = "foo 1234 1425059762"
	m := new(amproxy.Message)
	err := m.Decompose("foo 1234 1425059762 my_public_key lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=")
	if err != nil {
		t.Errorf("Shouldn't have gotten an error")
	}

	got := m.MetricStr()

	if got != expected {
		t.Errorf("MetricStr expected: %q got: %q", expected, got)
	}
}
