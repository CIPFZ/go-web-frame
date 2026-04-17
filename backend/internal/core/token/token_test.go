package token

import "testing"

func TestGenerateRawTokenUsesCmsPrefix(t *testing.T) {
	raw := GenerateRawToken()
	if raw == "" {
		t.Fatal("GenerateRawToken() returned empty token")
	}
	if len(raw) <= len(TokenPrefix) {
		t.Fatalf("GenerateRawToken() = %q, want token content after prefix", raw)
	}
	if raw[:len(TokenPrefix)] != TokenPrefix {
		t.Fatalf("GenerateRawToken() prefix = %q, want %q", raw[:len(TokenPrefix)], TokenPrefix)
	}
}

func TestHashTokenIsStable(t *testing.T) {
	const raw = "cms_test_token"

	hash1 := HashToken(raw)
	hash2 := HashToken(raw)

	if hash1 == "" || hash2 == "" {
		t.Fatal("HashToken() returned empty hash")
	}
	if hash1 != hash2 {
		t.Fatalf("HashToken() hashes differ: %q != %q", hash1, hash2)
	}
}

func TestVerifyTokenMatchesHash(t *testing.T) {
	const raw = "cms_verify_token"
	hash := HashToken(raw)

	if !VerifyToken(raw, hash) {
		t.Fatal("VerifyToken() = false, want true")
	}
	if VerifyToken(raw+"x", hash) {
		t.Fatal("VerifyToken() = true for mismatched token, want false")
	}
}
