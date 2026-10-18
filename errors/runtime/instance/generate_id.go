package instance

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

const (
	bufferLen   = 12 // разрядов base36 на число (метка времени и случайное число); в ID попадает срез [idOffset:idOffset+idBufferLen]
	idOffset    = 1  // ведущие обрезанные символы
	idDash1Pos  = 11
	idDash2Pos  = 16
	idBufferLen = 20 // timestamp(10) + dash(1) + random1(4) + dash(1) + random2(4)
)

//nolint:gochecknoglobals
var (
	baseNChars    = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")        // символы для кодирования Base-N
	baseNCharsLen = uint64(len(baseNChars))                               // количество символов для кодирования
	rndIfError    = [...]byte{0x0, 0x0, 0xee, 0xee, 0xee, 0xee, 0x0, 0x0} // данные по умолчанию при ошибке
)

// GenerateID - генерирует идентификатор в формате TTTTTTTTTT-XXXX-XXXX (символы 0-9, A-Z):
// 10 символов метки времени (UnixNano в base36 без старшего и младшего разрядов: шаг 36 нс,
// значение повторяется примерно каждые 4 года), далее 8 символов случайного числа.
// В пределах одного шага метки времени (и при её повторе через ~4 года) уникальность
// обеспечивают только 8 случайных символов (36^8 ≈ 2.8·10^12 значений): этого достаточно
// для сопоставления ошибки с записью в логе, но не для глобально уникального ключа.
// Используется для идентификации конкретных экземпляров runtime-ошибок.
func GenerateID() string {
	var (
		buf [bufferLen * 2]byte // (timestamp + random)
		rnd [len(rndIfError)]byte
	)

	encodeBaseN(buf[:bufferLen], uint64(time.Now().UnixNano()))

	if _, err := rand.Read(rnd[:]); err != nil {
		rnd = rndIfError
	}

	// смещение от взятия по модулю 36^12 затрагивает только старший разряд (~0.005 бита) и для ID несущественно
	encodeBaseN(buf[bufferLen:], binary.BigEndian.Uint64(rnd[:]))

	buf[idDash1Pos] = '-'
	buf[idDash2Pos] = '-'

	return string(buf[idOffset : idOffset+idBufferLen]) // TTTTTTTTTT-XXXX-XXXX
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
