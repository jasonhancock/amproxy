package auth

import "testing"

func TestParse(t *testing.T) {

    data := Parse("user1:foo1,user2:foo2,user3:foo3:blah,user4")

    user1 := data["user1"]
    if user1 != "foo1" {
        t.Errorf("expected %s, got %s", "foo1", user1)
    }

    user2 := data["user2"]
    if user2 != "foo2" {
        t.Errorf("expected %s, got %s", "foo2", user2)
    }

    _, ok3 := data["user3"]
    if ok3 {
        t.Errorf("Didn't expect user3 to be set")
    }

    _, ok4 := data["user4"]
    if ok4 {
        t.Errorf("Didn't expect user4 to be set")
    }
}
