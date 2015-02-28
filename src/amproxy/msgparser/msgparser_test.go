package msgparser

import "testing"

func TestDecompose(t *testing.T) {

    _, err := Decompose("foo 1234 1425059762 my_public_key")
    if err == nil {
        t.Errorf("invalid num of components. Expected error, didn't get it")
    }

    got, err := Decompose("foo 1234 1425059762 my_public_key lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=")
    //Decompose("foo 1234 1425059762 my_public_key")

    if got.Name != "foo" {
        t.Errorf("Name expected: %q got: %q", "foo", got.Name)
    }

    if got.Value != 1234 {
        t.Errorf("Value expected: %d got: %d", 1234, got.Value)
    }

    if got.Timestamp != 1425059762 {
        t.Errorf("Timestamp expected: %d got: %d", 1425059762, got.Timestamp)
    }

    if got.Public_key != "my_public_key" {
        t.Errorf("Public_key expected: %q got: %q", "my_public_key", got.Public_key)
    }

    if got.Signature != "lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=" {
        t.Errorf("Signature expected: %q got: %q", "lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=", got.Signature)
    }
}
