package backend

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

const signingKey = "abc123"
const email = "user@example.com"

func TestEncodeBytes(t *testing.T) {
	result := encodeBytes([]byte("abc"))
	if result != "K3K7K9" {
		t.Errorf("abc encoded should be K3K7K9, instead got: %s", result)
	}
}

func TestCheckVerifyCode(t *testing.T) {
	token := MakeVerifyCodeNow(signingKey, email)
	if !CheckVerifyCode(token, signingKey, email) {
		t.Errorf("Token %s did not validate", token)
	}

	if !CheckVerifyCode(strings.ToLower(token), signingKey, email) {
		t.Errorf("Token %s did not validate", strings.ToLower(token))
	}

	if CheckVerifyCode("ABC123", signingKey, email) {
		t.Error("Token ABC123 validated")
	}

	if CheckVerifyCode("ABC123", signingKey, email) {
		t.Error("Token ABC123 validated")
	}
}

func TestMakeVerifyCodeUnique(t *testing.T) {
	var tokens []string
	prevTime := time.Now().Add(-timeChunks)

	for i := 1; i <= 100; i += 1 {
		email := fmt.Sprintf("user%d@example.com", i)
		newToken := MakeVerifyCodeNow(signingKey, email)
		for _, oldToken := range tokens {
			if oldToken == newToken {
				t.Errorf("Email %s new token %s matched a previous one", email, newToken)
			}
		}

		tokens = append(tokens, newToken)

		prevToken := MakeVerifyCodeTS(signingKey, email, prevTime)
		for _, oldToken := range tokens {
			if oldToken == prevToken {
				t.Errorf("Email %s prev token %s matched a previous one", email, prevToken)
			}
		}

		tokens = append(tokens, prevToken)
	}
}
