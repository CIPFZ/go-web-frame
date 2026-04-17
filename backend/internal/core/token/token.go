package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

const (
	TokenPrefix = "cms_"
	TokenLength = 24
)

func GenerateRawToken() string {
	buf := make([]byte, TokenLength)
	_, _ = rand.Read(buf)
	return TokenPrefix + base64.RawURLEncoding.EncodeToString(buf)[:TokenLength]
}

func HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

func VerifyToken(rawToken, storedHash string) bool {
	return HashToken(rawToken) == storedHash
}
