// Package hash вычисляет и проверяет HMAC-подпись данных.
package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func ComputeHMAC(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	_, _ = h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func Verify(data []byte, key, receivedHash string) bool {
	expected := ComputeHMAC(data, key)
	return hmac.Equal([]byte(expected), []byte(receivedHash))
}
