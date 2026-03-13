package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Compress(data []byte) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(data)
	gz.Close()
	return buf.Bytes()
}

func ComputeHMAC(message []byte, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}
