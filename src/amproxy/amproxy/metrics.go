package main

import (
    "strings"
)

type counters struct {
    Connections uint64 `json:"connections"`
    BadSkew uint64 `json:"badskew"`
    BadSig uint64 `json:"badsig"`
    BadKeyundef uint64 `json:"badkeyundef"`
    BadDecompose uint64 `json:"baddecompose"`
    BadMetric uint64 `json:"badmetric"`
    BadCarbonconn uint64 `json:"badcarbonconn"`
    BadCarbonwrite uint64 `json:"badcarbonwrite"`
    GoodMetric uint64 `json:"goodmetric"`
}

func (c counters) init() {
    c.Connections = 0
    c.BadSkew = 0
    c.BadSig = 0
    c.BadKeyundef = 0
    c.BadDecompose = 0
    c.BadMetric = 0
    c.BadCarbonwrite = 0
    c.BadCarbonconn = 0
    c.GoodMetric = 0
}

func ReverseDelimitedString(str, delimiter string) string {

    pieces := strings.Split(str, delimiter)

    var rev []string

    for i := len(pieces)-1; i >= 0; i-- {
        rev = append(rev, pieces[i])
    }

    return strings.Join(rev, delimiter)
}
