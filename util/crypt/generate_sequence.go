package crypt

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
)

const (
	maxCharsetLen = 128
)

//nolint:gochecknoglobals
var (
	charsetDigit = []byte("0123456789")
	charsetHex   = []byte("0123456789abcdef")
	charsetToken = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

// GenerateDigits - генерирует криптографически случайную последовательность из цифр указанной длины.
func GenerateDigits(length int) (string, error) {
	return GenerateSequence(charsetDigit, length)
}

// GenerateHex - генерирует криптографически случайную последовательность из шестнадцатеричных цифр указанной длины.
func GenerateHex(length int) (string, error) {
	return GenerateSequence(charsetHex, length)
}

// GenerateToken - генерирует криптографически случайную последовательность указанной длины для уникальных токенов.
// Использует алфавитно-цифровой набор (a-z, A-Z, 0-9).
func GenerateToken(length int) (string, error) {
	return GenerateSequence(charsetToken, length)
}

// GenerateSequence - генерирует криптографически случайную последовательность из указанного набора ASCII-символов.
// Параметры:
//   - charset - набор допустимых символов (от 1 до 128);
//   - length - длина результата (минимум 1).
//
// Подробности в GenerateBytes.
func GenerateSequence(charset []byte, length int) (string, error) {
	s, err := GenerateBytes(charset, length)
	if err != nil {
		return "", err
	}

	return string(s), nil
}

// GenerateBytes - генерирует криптографически случайный слайс байтов из указанного набора символов.
// Символы выбираются равновероятно: случайные байты, выходящие за пределы charset, отбрасываются.
// Параметры:
//   - charset - набор допустимых символов (от 1 до 128);
//   - length - длина результата (минимум 1).
func GenerateBytes(charset []byte, length int) ([]byte, error) {
	if len(charset) == 0 {
		return nil, errors.New("charset is empty")
	}

	if len(charset) > maxCharsetLen {
		return nil, errors.New("charset length exceeds max length")
	}

	if length < 1 {
		return nil, errors.New("length less than 1")
	}

	chunk := make([]byte, length*2)
	sequence := chunk[:0]
	indexes := chunk[length:]

	bits100 := uint64(math.Log2(float64(len(charset))) * 100)
	bits := bits100 / 100

	if bits100%100 != 0 {
		bits++
	}

	mask := uint8(1<<bits) - 1

	for {
		if _, err := rand.Read(indexes); err != nil {
			return nil, fmt.Errorf("crypt.GenerateBytes: %w", err)
		}

		read := 0

		for i := 0; i < len(indexes); i++ {
			rnd := indexes[i] & mask

			if int(rnd) < len(charset) {
				sequence = append(sequence, charset[rnd])
				read++
			}
		}

		if read >= len(indexes) {
			return sequence[0:length:length], nil
		}

		indexes = indexes[:len(indexes)-read]
	}
}
