package auth

import (
    "strings"
)
// Given a comma-delimted string representing key-value pairs separated by a colon, return a map of key->value
// user1:pass1,user2:pass2
func Parse(str string) map[string]string {
    ret := make(map[string]string)
    pieces := strings.Split(str, ",")

    for _, value := range pieces {
        pieces2 := strings.Split(value, ":")
        if len(pieces2) == 2 {
            ret[pieces2[0]] = pieces2[1]
        }
    }

    return ret
}
