package amproxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrInvalidNumMessageComponents is the error returned when parsing a message
// that has an invaid number of items.
var ErrInvalidNumMessageComponents = errors.New("invalid number of components in message")

// Message represents an amproxy metric
type Message struct {
	Name      string
	Value     string
	Timestamp time.Time
	PublicKey string
	Signature string
}

// String returns the wire format of the message. The signature must have already been computed.
func (m Message) String() string {
	return fmt.Sprintf("%s %s %d %s %s", m.Name, m.Value, m.Timestamp.Unix(), m.PublicKey, m.Signature)
}

// Parse attempts to parse the given unsigned string into a Message.
func Parse(str string) (*Message, error) {
	return parse(str, false)
}

// ParseSigned attempts to parse the given signed string into a Message.
func ParseSigned(str string) (*Message, error) {
	return parse(str, true)
}

func parse(str string, signed bool) (*Message, error) {
	pieces := strings.Split(strings.TrimSpace(str), " ")
	expected := 5
	if !signed {
		expected = 4
	}
	if len(pieces) != expected {
		return nil, ErrInvalidNumMessageComponents
	}

	m := Message{
		Name:      pieces[0],
		Value:     pieces[1],
		PublicKey: pieces[3],
	}

	if signed {
		m.Signature = pieces[4]
	}

	unixTime, err := strconv.ParseInt(pieces[2], 10, 64)
	if err != nil {
		return nil, err
	}

	m.Timestamp = time.Unix(unixTime, 0)

	return &m, nil
}

// ComputeSignature uses the private key to compute the HMAC SHA-256 signature for the message.
func (m Message) ComputeSignature(secret string) string {
	message := fmt.Sprintf("%s %s %d %s", m.Name, m.Value, m.Timestamp.Unix(), m.PublicKey)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// MetricStr returns the carbon wire format of the message.
func (m Message) MetricStr() string {
	return fmt.Sprintf("%s %s %d", m.Name, m.Value, m.Timestamp.Unix())
}
