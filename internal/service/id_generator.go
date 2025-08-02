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

// IDGenerator는 Base62 인코딩을 사용한 고유 ID 생성기입니다
type IDGenerator struct {
	length int
}

// NewIDGenerator는 새로운 ID 생성기를 생성합니다
func NewIDGenerator(length int) *IDGenerator {
	if length < 3 {
		length = defaultIDLength
	}
	return &IDGenerator{
		length: length,
	}
}

// Generate는 랜덤한 Base62 ID를 생성합니다
func (g *IDGenerator) Generate() (string, error) {
	var result strings.Builder
	result.Grow(g.length)
	
	for i := 0; i < g.length; i++ {
		// 암호학적으로 안전한 랜덤 숫자 생성
		num, err := rand.Int(rand.Reader, big.NewInt(base62Base))
		if err != nil {
			return "", err
		}
		result.WriteByte(base62Chars[num.Int64()])
	}
	
	return result.String(), nil
}

// EncodeNumber는 숫자를 Base62로 인코딩합니다
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

// DecodeToNumber는 Base62 문자열을 숫자로 디코딩합니다
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

// IsValidID는 주어진 문자열이 유효한 Base62 ID인지 확인합니다
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

// GenerateWithPrefix는 접두사가 있는 ID를 생성합니다
func (g *IDGenerator) GenerateWithPrefix(prefix string) (string, error) {
	id, err := g.Generate()
	if err != nil {
		return "", err
	}
	return prefix + id, nil
}

// 유틸리티 함수들

// QuickGenerate는 기본 길이의 랜덤 ID를 빠르게 생성합니다
func QuickGenerate() (string, error) {
	generator := NewIDGenerator(defaultIDLength)
	return generator.Generate()
}

// QuickEncode는 숫자를 Base62로 빠르게 인코딩합니다
func QuickEncode(num int64) string {
	generator := NewIDGenerator(defaultIDLength)
	return generator.EncodeNumber(num)
}

// QuickDecode는 Base62 문자열을 숫자로 빠르게 디코딩합니다
func QuickDecode(encoded string) (int64, error) {
	generator := NewIDGenerator(defaultIDLength)
	return generator.DecodeToNumber(encoded)
}