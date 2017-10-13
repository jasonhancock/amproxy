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

// ErrInvalidNumMessageComponents is the error returned when parsing a message that has an invaid number of items
var ErrInvalidNumMessageComponents = errors.New("Invalid number of components in message")

// Message represents an amproxy metric
type Message struct {
	Name      string
	Value     string
	Timestamp int
	PublicKey string
	Signature string
}

// String returns the wire format of the message. The signature must have already been computed
func (m Message) String() string {
	return fmt.Sprintf("%s %s %d %s %s", m.Name, m.Value, m.Timestamp, m.PublicKey, m.Signature)
}

// Parse attempts to parse the given string into a Message
func Parse(str string) (*Message, error) {
	m := &Message{}
	pieces := strings.Split(strings.TrimSpace(str), " ")

	if len(pieces) != 5 {
		return nil, ErrInvalidNumMessageComponents
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

// ComputeSignature uses the private key to compute the HMAC SHA-256 signature for the message
func (m Message) ComputeSignature(secret string) string {
	message := fmt.Sprintf("%s %s %d %s", m.Name, m.Value, m.Timestamp, m.PublicKey)
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// MetricStr returns the carbon wire format of the message
func (m Message) MetricStr() string {
	return fmt.Sprintf("%s %s %d", m.Name, m.Value, m.Timestamp)
}
