package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"

	"github.com/sirupsen/logrus"
)

func RandToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// ByteToMap converts a stream of bytes to a map
func ByteToMap(data []byte) map[string]interface{} {
	res := map[string]interface{}{}
	if err := json.Unmarshal(data, &res); err != nil {
		logrus.Error(err)
	}
	return res
}
