package backend

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"strings"
	"time"
)

// Set of easily distinguishable characters
// const CODE_CHARS = "23579CDFHJKLMNPQRSTVWXYZ"

const CodeChars = "2379HJKLNQSTVXYZ" // Limited to a nice even 16 options
const timeChunks = time.Duration(15) * time.Minute
const hashLen = 8

// / encodeBytes Generates easy human-readable strings out of input bytes
func encodeBytes(input []byte) string {
	result := ""

	for _, value := range input {
		left := value >> 4
		right := value & 0b00001111
		valueStr := CodeChars[left:left+1] + CodeChars[right:right+1]
		result += valueStr
	}

	return result
}

// / getTimeChunk Gets a neat timestamp string of the chunk of time the given time belongs to
func getTimeChunk(t time.Time) string {
	// Truncate to chunk
	return t.Truncate(timeChunks).String()
}

// / getHash Produce a strong hash unique to the signing key and value
func getHash(signingKey string, value string) []byte {
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(value))
	hash := mac.Sum(nil)
	return hash
}

func MakeVerifyCodeTS(signingKey string, email string, timestamp time.Time) string {
	payload := fmt.Sprintf("%s | %s", email, getTimeChunk(timestamp))
	hash := getHash(signingKey, payload)
	return encodeBytes(hash[:hashLen])[:hashLen]
}

func MakeVerifyCodeNow(signingKey string, email string) string {
	return MakeVerifyCodeTS(signingKey, email, time.Now())
}

func CheckVerifyCodeTS(token string, signingKey string, email string, timestamp time.Time) bool {
	token = strings.ToUpper(token)
	currentExpected := MakeVerifyCodeTS(signingKey, email, timestamp)
	previousExpected := MakeVerifyCodeTS(signingKey, email, timestamp.Add(-timeChunks))

	match := false

	if subtle.ConstantTimeCompare([]byte(token), []byte(currentExpected)) == 1 {
		match = true
	}

	if subtle.ConstantTimeCompare([]byte(token), []byte(previousExpected)) == 1 {
		match = true
	}

	return match
}

func CheckVerifyCode(token string, signingKey string, email string) bool {
	return CheckVerifyCodeTS(token, signingKey, email, time.Now())
}
