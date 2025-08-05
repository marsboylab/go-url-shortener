package service

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	// Base62 문자 집합: 0-9, a-z, A-Z (URL 안전)
	base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base62Base  = int64(len(base62Chars))
	
	// 기본 ID 길이
	defaultIDLength = 6
)

type IDGenerator struct {
	length int
}

func NewIDGenerator(length int) *IDGenerator {
	if length < 3 {
		length = defaultIDLength
	}
	return &IDGenerator{
		length: length,
	}
}

func (g *IDGenerator) Generate() (string, error) {
	var result strings.Builder
	result.Grow(g.length)
	
	for i := 0; i < g.length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(base62Base))
		if err != nil {
			return "", err
		}
		result.WriteByte(base62Chars[num.Int64()])
	}
	
	return result.String(), nil
}

func (g *IDGenerator) EncodeNumber(num int64) string {
	if num == 0 {
		return "0"
	}
	
	var result strings.Builder
	
	for num > 0 {
		remainder := num % base62Base
		result.WriteByte(base62Chars[remainder])
		num = num / base62Base
	}
	
	// 문자열 뒤집기
	encoded := result.String()
	runes := []rune(encoded)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	
	return string(runes)
}

func (g *IDGenerator) DecodeToNumber(encoded string) (int64, error) {
	var result int64
	var power int64 = 1
	
	// 문자열을 뒤에서부터 처리
	for i := len(encoded) - 1; i >= 0; i-- {
		char := encoded[i]
		index := strings.IndexByte(base62Chars, char)
		if index == -1 {
			return 0, NewValidationError("decode_error", "Invalid character in Base62 string", map[string]interface{}{
				"character": string(char),
				"position":  len(encoded) - 1 - i,
			})
		}
		
		result += int64(index) * power
		power *= base62Base
	}
	
	return result, nil
}

func (g *IDGenerator) IsValidID(id string) bool {
	if len(id) == 0 {
		return false
	}
	
	for _, char := range id {
		if !strings.ContainsRune(base62Chars, char) {
			return false
		}
	}
	
	return true
}

func (g *IDGenerator) GenerateWithPrefix(prefix string) (string, error) {
	id, err := g.Generate()
	if err != nil {
		return "", err
	}
	return prefix + id, nil
}

// utility functions
func QuickGenerate() (string, error) {
	generator := NewIDGenerator(defaultIDLength)
	return generator.Generate()
}

func QuickEncode(num int64) string {
	generator := NewIDGenerator(defaultIDLength)
	return generator.EncodeNumber(num)
}

func QuickDecode(encoded string) (int64, error) {
	generator := NewIDGenerator(defaultIDLength)
	return generator.DecodeToNumber(encoded)
}