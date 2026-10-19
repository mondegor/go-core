package mutexlocker

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/mondegor/go-core/mrlock"
)

const (
	// lockerName - имя блокировщика для логирования и трассировки.
	lockerName = "MutexLocker"

	// defaultMinBufferSize - минимальный размер буфера хранения ключей.
	defaultMinBufferSize = 32

	// Какое максимальное кол-во протухших ключей удалять за один раз.
	removeExpiredKeysLimit = 4
)

type (
	// Locker - реализация интерфейса блокировщика ключа на основе sync.Mutex.
	// Блокировка действует только в рамках текущего процесса приложения.
	// Хранит заблокированные ключи в map с временными метками истечения.
	Locker struct {
		logger logger           // logger - логгер для предупреждений
		tracer tracer           // tracer - трассировщик для логирования операций
		mu     sync.Mutex       // mu - мьютекс для синхронизации доступа к map keys
		keys   map[string]int64 // keys - карта заблокированных ключей с timestamp истечения в наносекундах
	}

	logger interface {
		Warn(ctx context.Context, msg string, args ...any)
	}

	tracer interface {
		Trace(ctx context.Context, args ...any)
	}
)

// New - создаёт объект Locker.
func New(logger logger, tracer tracer) *Locker {
	return &Locker{
		logger: logger,
		tracer: tracer,
		keys:   make(map[string]int64, defaultMinBufferSize),
	}
}

// Lock - блокирует указанный ключ в рамках приложения с временем блокировки по умолчанию
// и возвращает функцию разблокировки.
func (l *Locker) Lock(ctx context.Context, key string) (unlock func(), err error) {
	return l.LockWithExpiry(ctx, key, 0)
}

// LockWithExpiry - блокирует указанный ключ в рамках приложения с указанием
// времени блокировки и возвращает функцию разблокировки.
// Если указана expiry равная нулю, то будет установлено время блокировки по умолчанию.
func (l *Locker) LockWithExpiry(ctx context.Context, key string, expiry time.Duration) (unlock func(), err error) {
	if expiry == 0 {
		expiry = mrlock.DefaultExpiry
	}

	l.traceCmd(ctx, "Lock:"+expiry.String()+", keys-len="+strconv.Itoa(len(l.keys)), key)

	l.mu.Lock()
	defer l.mu.Unlock()

	l.keysAutoCleaner(ctx, removeExpiredKeysLimit)

	if exp, ok := l.keys[key]; ok && exp > time.Now().UnixNano() {
		return nil, fmt.Errorf(
			"%w [source=%s, lock_key=%q]",
			mrlock.ErrLockKeyNotObtained, lockerName, key,
		)
	}

	l.keys[key] = time.Now().UnixNano() + expiry.Nanoseconds()

	return func() {
		l.traceCmd(ctx, "Unlock", key)

		l.mu.Lock()
		defer l.mu.Unlock()

		if _, ok := l.keys[key]; !ok {
			l.logger.Warn(
				ctx,
				"unlock",
				"error", mrlock.ErrLockKeyNotHeld,
				"source", lockerName,
				"lock_key", key,
			)

			return
		}

		delete(l.keys, key)
	}, nil
}

// traceCmd - логирует выполняемую операцию блокировки для трассировки.
func (l *Locker) traceCmd(ctx context.Context, command, key string) {
	l.tracer.Trace(
		ctx,
		"source", lockerName,
		"cmd", command,
		"key", key,
	)
}

// keysAutoCleaner - проверяет autoCleaner протухших ключей и удаляет их.
// Актуально, только если кто-то забыл вызвать unlock().
func (l *Locker) keysAutoCleaner(ctx context.Context, limit int) {
	curTime := time.Now().UnixNano()

	for key, exp := range l.keys {
		if exp > curTime {
			continue
		}

		delete(l.keys, key)

		limit--

		if limit > 0 {
			continue
		}

		// по этому логу можно будет выяснить, что кто-то забыл вызывать unlock().
		l.logger.Warn(
			ctx,
			"keysAutoCleaner",
			"source", lockerName,
			"last_deleted_key", key,
		)

		break
	}
}
