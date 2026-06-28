package shortcode

import (
	"crypto/rand"
	"math/big"
)

// alphabet is base62 — a-z, A-Z, 0-9.
// 62^6 = ~56 billion possible codes, plenty for a URL shortener.
const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Generate returns a cryptographically random 6-character base62 short code.
// crypto/rand is used instead of math/rand because codes must be unpredictable —
// a predictable RNG would let an attacker enumerate all shortened links.
func Generate() (string, error) {
	const length = 6
	result := make([]byte, length)
	for i := range result {
		// Pick a random index into the alphabet
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		result[i] = alphabet[n.Int64()]
	}
	return string(result), nil
}
