package amproxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var errInvalidNumMessageComponents = errors.New("Invalid number of components in message")

type Message struct {
	Name      string
	Value     string
	Timestamp int
	PublicKey string
	Signature string
}

func (m Message) String() string {
	return fmt.Sprintf("%s %d %d %s %s", m.Name, m.Value, m.Timestamp, m.PublicKey, m.Signature)
}

func Decompose(str string) (*Message, error) {
	m := &Message{}
	pieces := strings.Split(strings.TrimSpace(str), " ")

	if len(pieces) != 5 {
		return nil, errInvalidNumMessageComponents
	}

	m.Name = pieces[0]
	m.Value = pieces[1]

	timestamp, err := strconv.Atoi(pieces[2])
	if err != nil {
		return nil, err
	}
	m.Timestamp = timestamp

	m.PublicKey = pieces[3]
	m.Signature = pieces[4]

	return m, nil
}

func (m Message) ComputeSignature(secret string) string {
	message := fmt.Sprintf("%s %s %d %s", m.Name, m.Value, m.Timestamp, m.PublicKey)
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (m Message) MetricStr() string {
	return fmt.Sprintf("%s %s %d", m.Name, m.Value, m.Timestamp)
}
