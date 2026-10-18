package mrtrace

import (
	"context"
)

type (
	// nopTracer - заглушка, реализующая интерфейс Tracer.
	// Игнорирует все данные трассировки.
	nopTracer struct{}
)

// NopTracer - создаёт Tracer, который игнорирует все данные трассировки.
func NopTracer() Tracer {
	return nopTracer{}
}

// Trace - игнорирует данные трассировки (заглушка для интерфейса Tracer).
func (t nopTracer) Trace(_ context.Context, _ ...any) {}
