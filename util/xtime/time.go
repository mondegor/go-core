package xtime

import (
	"time"
)

// TimeLeftInSec - возвращает количество секунд, оставшихся до указанного времени,
// округлённое до ближайшей целой секунды. Для прошедшего времени возвращается 0.
func TimeLeftInSec(tm time.Time) int64 {
	const (
		second     = 1e9
		halfSecond = second / 2
	)

	rest := tm.UnixNano() - time.Now().UnixNano() + halfSecond

	if rest < 0 {
		return 0
	}

	return rest / second
}
