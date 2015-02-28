package msgparser

import (
    "fmt"
    "strconv"
    "strings"
)

type Message struct {
    Name string
    Value int
    Timestamp int
    Public_key string
    Signature string
}

func Decompose(str string) (*Message, error) {
    msg := new(Message)
    pieces := strings.Split(strings.TrimSpace(str), " ")

    length := len(pieces)
    if(length != 5) {
        return msg, fmt.Errorf("msgparser: invalid number of components: %d", length)
    }

    msg.Name = pieces[0]

    value, err := strconv.Atoi(pieces[1])
    if err != nil {
        return msg, fmt.Errorf("Error parsing metric value: %q", pieces[1])
    }
    msg.Value = value

    timestamp, err2 := strconv.Atoi(pieces[2])
    if err2 != nil {
        return msg, fmt.Errorf("Error parsing timestamp value: %q", pieces[2])
    }
    msg.Timestamp = timestamp

    msg.Public_key = pieces[3]
    msg.Signature  = pieces[4]

    return msg, nil
}
