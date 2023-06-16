package random

import (
	"crypto/rand"
	"encoding/hex"
)

const (
	whitespace    = 32
	quotationMark = 34
	backslash     = 92
	lessThanSign  = 60
	delete        = 127
)

// https://gist.github.com/denisbrodbeck/635a644089868a51eccd6ae22b2eb800 (source)
func ascii(length int) (string, error) {
	result := make([]byte, length)
	if _, err := rand.Read(result); err != nil {
		return "", err
	}
	for i := 0; i < length; i++ {
		result[i] &= 0x7F
		for result[i] <= whitespace || result[i] == delete || result[i] == quotationMark || result[i] == backslash || result[i] == lessThanSign {
			if _, err := rand.Read(result[i : i+1]); err != nil {
				return "", err
			}
			result[i] &= 0x7F
		}
	}
	return string(result), nil
}

func ID(len int) string {
	id, _ := ascii(len)
	return hex.EncodeToString([]byte(id))
}

func String(len int) string {
	str, _ := ascii(len)
	return str
}
