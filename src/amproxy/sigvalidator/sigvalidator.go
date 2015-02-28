package sigvalidator

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
)

// Computes Base64 encoded SHA256 signature
func ComputeSignature(message string, secret string) string {
    key := []byte(secret)
    h := hmac.New(sha256.New, key)
    h.Write([]byte(message))
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
