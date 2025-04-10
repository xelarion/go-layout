package util

import (
	"math/rand"
)

var numberRunes = []rune("0123456789")
var characterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandNumberSeq 长度为 length 的随机数字序列
func RandNumberSeq(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = numberRunes[rand.Intn(len(numberRunes))]
	}
	return string(b)
}

// RandSeq 长度为 length 的随机字母数字序列
func RandSeq(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = characterRunes[rand.Intn(len(characterRunes))]
	}
	return string(b)
}
