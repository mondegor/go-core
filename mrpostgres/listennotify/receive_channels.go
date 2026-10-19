package listennotify

import (
	"fmt"
)

type (
	// receiveChannel - канал для получения уведомлений от PostgreSQL.
	// Содержит имя канала и сам канал для передачи сигналов.
	receiveChannel struct {
		Name    string          // Name - имя канала подписки PostgreSQL
		Channel <-chan struct{} // Channel - канал для получения уведомлений
	}

	// receiveChannels - коллекция каналов для получения уведомлений от PostgreSQL.
	receiveChannels []receiveChannel
)

// Find - находит канал по имени и возвращает его для получения уведомлений.
func (rc *receiveChannels) Find(name string) (<-chan struct{}, error) {
	for _, rch := range *rc {
		if name == rch.Name {
			return rch.Channel, nil
		}
	}

	return nil, fmt.Errorf("no such channel (name=%q)", name)
}
