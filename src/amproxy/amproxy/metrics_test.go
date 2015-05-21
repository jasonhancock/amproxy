package main

import (
    "testing"
)


func TestReverseString(t *testing.T) {
    if ReverseDelimitedString("foo.bar.baz", ".") != "baz.bar.foo" {
        t.Errorf("reverse delimited string failed")
    }
}
