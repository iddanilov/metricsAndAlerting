package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

func CreateHash(m string, key []byte) (string, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(m))
	return fmt.Sprintf("%x", h.Sum(nil)), err
}
