package message

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "fmt"
    "strconv"
    "strings"
)

type Message struct {
    Name string
    Value string
    Timestamp int
    Public_key string
    Signature string
}

func (m Message) String() string {
    return fmt.Sprintf("%s %d %d %s %s", m.Name, m.Value, m.Timestamp, m.Public_key, m.Signature)
}

func (m *Message) Decompose(str string) error {
    pieces := strings.Split(strings.TrimSpace(str), " ")

    length := len(pieces)
    if(length != 5) {
        return fmt.Errorf("message: invalid number of components: %d", length)
    }

    m.Name = pieces[0]
    m.Value = pieces[1]

    timestamp, err2 := strconv.Atoi(pieces[2])
    if err2 != nil {
        return fmt.Errorf("Error parsing timestamp value: %q", pieces[2])
    }
    m.Timestamp = timestamp

    m.Public_key = pieces[3]
    m.Signature  = pieces[4]

    return nil
}

func (m Message) ComputeSignature(secret string) string {
    message := fmt.Sprintf("%s %s %d %s", m.Name, m.Value, m.Timestamp, m.Public_key)
    key := []byte(secret)
    h := hmac.New(sha256.New, key)
    h.Write([]byte(message))
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (m Message) MetricStr() string {
    return fmt.Sprintf("%s %s %d", m.Name, m.Value, m.Timestamp)
}
