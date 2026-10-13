package xio

import (
	"context"
	"fmt"
	"io"
)

type (
	// CloseManager - менеджер для закрытия всех зарегистрированных ресурсов.
	// Не потокобезопасен: Register и Close вызываются из одной горутины.
	CloseManager struct {
		logger    logger
		resources []io.Closer
	}
)

// NewCloseManager - создаёт менеджер для закрытия зарегистрированных ресурсов.
// Параметр logger используется для логирования успешного закрытия и ошибок.
func NewCloseManager(logger logger) *CloseManager {
	return &CloseManager{
		logger: logger,
	}
}

// Register - регистрирует указанный ресурс в своём списке.
func (m *CloseManager) Register(resource io.Closer) {
	m.resources = append(m.resources, resource)

	m.logger.Debug(context.Background(), fmt.Sprintf("Resource %T was registered", resource))
}

// Close - закрывает ресурсы в порядке регистрации. Ошибка закрытия ресурса
// логируется и не прерывает закрытие остальных.
func (m *CloseManager) Close() {
	for _, resource := range m.resources {
		m.close(resource)
	}
}

func (m *CloseManager) close(resource io.Closer) {
	if err := resource.Close(); err != nil {
		m.logger.Error(
			context.Background(),
			"CloseManager.Close(): failed to close object",
			"error", err,
			"resource", fmt.Sprintf("%#v", resource),
		)

		return
	}

	m.logger.Debug(context.Background(), fmt.Sprintf("Resource %T was closed", resource))
}
