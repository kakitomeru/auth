package crypto

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRandomBase64(n int) string {
	bytes := make([]byte, n)
	rand.Read(bytes)

	return base64.URLEncoding.EncodeToString(bytes)
}
