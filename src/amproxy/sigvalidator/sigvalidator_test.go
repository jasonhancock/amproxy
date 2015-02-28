package sigvalidator

import "testing"

func TestComputeSignature(t *testing.T) {

    var expected string = "lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc="
    got := ComputeSignature("foo 1234 1425059762 my_public_key", "my_secret_key")

    if got != expected {
        t.Errorf("ComputeSignature expected: %q got: %q", expected, got)
    }
}
