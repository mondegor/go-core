package process

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

const (
	bufferLen   = 12 // 36^12 ns ≈ 150 лет: хватает для UnixNano до ~2120 года; у 64-битного случайного числа старший разряд отбрасывается
	idDash1Pos  = 12
	idDash2Pos  = 19
	idBufferLen = 24 // timestamp(12) + dash(1) + random1(6) + dash(1) + random2(4)
)

//nolint:gochecknoglobals
var (
	baseNChars    = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")        // символы для кодирования Base-N
	baseNCharsLen = uint64(len(baseNChars))                               // количество символов для кодирования
	rndIfError    = [...]byte{0x0, 0x0, 0xee, 0xee, 0xee, 0xee, 0x0, 0x0} // данные по умолчанию при ошибке
)

// GenerateID - генерирует идентификатор в формате TTTTTTTTTTTT-XXXXXX-XXXX (символы 0-9, A-Z):
// 12 символов метки времени (UnixNano в base36), далее 10 символов случайного числа.
// Используется для идентификации процессов при трассировке запросов.
func GenerateID() string {
	var (
		buf [bufferLen * 2]byte // (timestamp + random)
		rnd [len(rndIfError)]byte
	)

	encodeBaseN(buf[:bufferLen], uint64(time.Now().UnixNano()))

	if _, err := rand.Read(rnd[:]); err != nil {
		rnd = rndIfError
	}

	encodeBaseN(buf[bufferLen:], binary.BigEndian.Uint64(rnd[:]))

	buf[idDash1Pos] = '-'
	buf[idDash2Pos] = '-'

	return string(buf[:idBufferLen]) // TTTTTTTTTTTT-XXXXXX-XXXX
}

// encodeBaseN - записывает в buf ровно bufferLen младших разрядов числа n в Base-N:
// недостающие старшие разряды заполняются символом '0', лишние старшие разряды отбрасываются.
func encodeBaseN(buf []byte, n uint64) {
	// цифры записываются с конца буфера, от младшего разряда к старшему
	for i := bufferLen - 1; i >= 0; i-- {
		q := n / baseNCharsLen
		buf[i] = baseNChars[n-q*baseNCharsLen]
		n = q
	}
}
