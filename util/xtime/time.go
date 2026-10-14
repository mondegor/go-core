package xtime

import (
	"time"
)

// TimeLeftInSec - возвращает количество секунд, оставшихся до указанного времени,
// округлённое в большую сторону. Для наступившего или прошедшего времени возвращается 0.
func TimeLeftInSec(tm time.Time) int64 {
	const second = int64(time.Second)

	rest := tm.UnixNano() - time.Now().UnixNano()

	if rest <= 0 {
		return 0
	}

	return (rest + second - 1) / second
}
