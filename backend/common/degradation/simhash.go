package degradation

import (
	"hash/fnv"
	"math/bits"
	"strings"
	"unicode"
)

const simHashShingleSize = 3

// ComputeSimHash calculates a 64-bit SimHash fingerprint for moderation use.
func ComputeSimHash(text string) uint64 {
	shingles := buildSimHashShingles(text)
	if len(shingles) == 0 {
		return 0
	}

	weights := make(map[string]int, len(shingles))
	for _, shingle := range shingles {
		weights[shingle]++
	}

	var vector [64]int
	for shingle, weight := range weights {
		hash := hashSimHashToken(shingle)
		for bit := 0; bit < 64; bit++ {
			if hash&(uint64(1)<<bit) != 0 {
				vector[bit] += weight
			} else {
				vector[bit] -= weight
			}
		}
	}

	var fingerprint uint64
	for bit, value := range vector {
		if value >= 0 {
			fingerprint |= uint64(1) << bit
		}
	}
	return fingerprint
}

func HammingDistance(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}

func buildSimHashShingles(text string) []string {
	tokens := tokenizeSimHash(text)
	if len(tokens) == 0 {
		return nil
	}

	size := simHashShingleSize
	if len(tokens) < size {
		size = len(tokens)
	}

	shingles := make([]string, 0, len(tokens)-size+1)
	for i := 0; i+size <= len(tokens); i++ {
		shingles = append(shingles, strings.Join(tokens[i:i+size], "|"))
	}
	return shingles
}

func tokenizeSimHash(text string) []string {
	var tokens []string
	var asciiBuilder strings.Builder

	flushASCII := func() {
		if asciiBuilder.Len() == 0 {
			return
		}
		tokens = append(tokens, asciiBuilder.String())
		asciiBuilder.Reset()
	}

	for _, r := range strings.ToLower(text) {
		switch {
		case isASCIIAlphaNum(r):
			asciiBuilder.WriteRune(r)
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			flushASCII()
			tokens = append(tokens, string(r))
		default:
			flushASCII()
		}
	}
	flushASCII()
	return tokens
}

func hashSimHashToken(token string) uint64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(token))
	return hasher.Sum64()
}

func isASCIIAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}
