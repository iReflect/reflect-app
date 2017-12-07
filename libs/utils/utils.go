package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func RandToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
