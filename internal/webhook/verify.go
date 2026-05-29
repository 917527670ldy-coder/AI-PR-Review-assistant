package webhook

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "strings"
)

func VerifySignature(secret string, body []byte, signature256 string) bool {
    if secret == "" || signature256 == "" {
        return false
    }

    const prefix = "sha256="
    if !strings.HasPrefix(signature256, prefix) {
        return false
    }

    gotHex := strings.TrimPrefix(signature256, prefix)
    got, err := hex.DecodeString(gotHex)
    if err != nil {
        return false
    }

    mac := hmac.New(sha256.New, []byte(secret))
    _, _ = mac.Write(body)
    expected := mac.Sum(nil)

    return hmac.Equal(got, expected)
}
