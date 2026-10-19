package listennotify

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/mondegor/go-core/mrlog"
	"github.com/mondegor/go-core/mrpostgres"
)

const (
	defaultReadyTimeout    = 5 * time.Second
	defaultReconnectDelay  = 5 * time.Second
	defaultCheckConnPeriod = time.Minute
)

type (
	// ProcessWaitForNotification - процесс прослушивания и обработки событий (NOTIFY) от PostgreSQL.
	// Переподключается к БД при разрыве соединения с настраиваемой задержкой.
	ProcessWaitForNotification struct {
		conn               *mrpostgres.ConnAdapter
		logger             mrlog.Logger
		listenerChannelMap map[string]chan struct{} // listenerChannelMap - маппинг имён каналов на каналы уведомлений
		reconnectDelay     time.Duration            // reconnectDelay - задержка между попытками переподключения
		checkConnPeriod    time.Duration            // checkConnPeriod - период, через которое будет проверяться активность соединения

		wg   sync.WaitGroup
		done chan struct{}

		receiverChannels receiveChannels // receiverChannels - коллекция каналов для подписчиков
	}
)

// NewProcessWaitForNotification - создаёт объект ProcessWaitForNotification для прослушивания NOTIFY от PostgreSQL.
// Параметры:
//   - conn - адаптер подключения к PostgreSQL;
//   - logger - логгер для вывода сообщений;
//   - channels - список имён каналов для подписки.
func NewProcessWaitForNotification(
	conn *mrpostgres.ConnAdapter,
	logger mrlog.Logger,
	channels []string,
	checkConnPeriod time.Duration,
) *ProcessWaitForNotification {
	listenerChannelMap, receiverChannelList := createListenerChannels(logger, channels)

	if checkConnPeriod < 1 {
		checkConnPeriod = defaultCheckConnPeriod
	}

	return &ProcessWaitForNotification{
		conn:               conn,
		logger:             logger,
		listenerChannelMap: listenerChannelMap,
		reconnectDelay:     defaultReconnectDelay,
		checkConnPeriod:    checkConnPeriod,

		wg:   sync.WaitGroup{},
		done: make(chan struct{}),

		receiverChannels: receiverChannelList,
	}
}

func createListenerChannels(logger mrlog.Logger, channels []string) (map[string]chan struct{}, []receiveChannel) {
	listenerChannels := make(map[string]chan struct{}, len(channels))
	receiveChannelList := make([]receiveChannel, 0, len(channels))

	for _, name := range channels {
		if _, ok := listenerChannels[name]; ok {
			mrlog.Warn(logger, "Duplicate listen channel", "channel", name)

			continue
		}

		channel := make(chan struct{})

		listenerChannels[name] = channel
		receiveChannelList = append(
			receiveChannelList,
			receiveChannel{
				Name:    name,
				Channel: channel,
			},
		)
	}

	return listenerChannels, receiveChannelList
}

// Caption - возвращает название процесса в свободной форме.
func (p *ProcessWaitForNotification) Caption() string {
	return "ProcessWaitForNotification"
}

// ReadyTimeout - возвращает таймаут готовности процесса для ожидания запуска.
func (p *ProcessWaitForNotification) ReadyTimeout() time.Duration {
	return defaultReadyTimeout
}

// Start - запускает процесс прослушивания NOTIFY от PostgreSQL.
// Блокирует выполнение до завершения контекста или возникновения ошибки.
// Автоматически переподключается при разрыве соединения с задержкой defaultReconnectDelay.
func (p *ProcessWaitForNotification) Start(ctx context.Context, ready func()) error {
	p.wg.Add(1)
	defer p.wg.Done()

	p.logger.Debug(ctx, "Starting the WaitForNotification...")
	defer p.logger.Debug(ctx, "The WaitForNotification has been stopped")

	ctxListen, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-p.done:
			cancel()
		case <-ctx.Done():
		}
	}()

	if ready != nil {
		ready()
	}

	for {
		if err := p.listen(ctxListen); err != nil {
			// если ошибка вызвана текущим контекстом, то
			// это просто завершается работа процесса
			if errors.Is(err, ctxListen.Err()) {
				return nil
			}

			p.logger.Error(ctxListen, "ProcessWaitForNotification.listen", "error", err)
		} else {
			return nil
		}

		if p.reconnectDelay < 1 {
			continue
		}

		select {
		case <-p.done:
			return nil
		case <-ctx.Done():
			p.logger.Debug(ctx, "The WaitForNotification detected context 'Done'", "error", ctx.Err())

			return nil
		case <-time.After(p.reconnectDelay):
		}
	}
}

func (p *ProcessWaitForNotification) listen(ctx context.Context) error {
	conn, err := p.conn.HijackConn(ctx)
	if err != nil {
		return fmt.Errorf("listen connect: %w", err)
	}

	defer func() {
		_ = conn.Close(ctx)
	}()

	for name := range p.listenerChannelMap {
		if _, err := conn.Exec(ctx, "LISTEN "+pgx.Identifier{name}.Sanitize()); err != nil {
			return fmt.Errorf("unable to start listening channel %q: %w", name, err)
		}
	}

	for {
		// если контекст завершен, то это значит, что сервис завершает работу (Shutdown)
		if ctx.Err() != nil {
			return ctx.Err()
		}

		p.logger.Debug(ctx, "Waiting for the notification or event timeout...")

		if err = p.waitSomePeriod(ctx, conn); err != nil {
			// если это отмена внутреннего контекста, то работа продолжится,
			// иначе процесс будет завершен при проверке `ctx.Err()` выше
			if errors.Is(err, context.Canceled) {
				continue
			}

			return fmt.Errorf("listen process error: %w", err)
		}
	}
}

func (p *ProcessWaitForNotification) waitSomePeriod(ctx context.Context, conn *pgx.Conn) error {
	errChan := make(chan error)

	// создаётся контекст, для возможности прерывания WaitForNotification по таймауту
	waitCtx, waitCancel := context.WithCancel(ctx)
	defer waitCancel()

	go func() {
		for {
			note, err := conn.WaitForNotification(waitCtx)
			if err != nil {
				errChan <- fmt.Errorf("conn.WaitForNotification: %w", err)

				return
			}

			if ch, ok := p.listenerChannelMap[note.Channel]; ok {
				// если канал занят, значит такое же событие ещё не обработано получателем,
				// поэтому нет смысла отправлять повторное событие, поэтому оно пропускается
				select {
				case ch <- struct{}{}:
					p.logger.Debug(ctx, fmt.Sprintf("Received notification: PID=%d, Channel=%q, Payload=%q", note.PID, note.Channel, note.Payload))
				default:
					p.logger.Info(ctx, fmt.Sprintf("Repeated notification: PID=%d, Channel=%q, Payload=%q [skipped]", note.PID, note.Channel, note.Payload))
				}

				continue
			}

			p.logger.Warn(
				ctx,
				"Unknown channel",
				"pid", note.PID,
				"channel", note.Channel,
				"payload", note.Payload,
			)
		}
	}()

	// функция вызывается при срабатывании таймаута для очередной проверки соединения
	// с целью прервать WaitForNotification и перехватить от него ошибку отмены контекста
	interruptWaitForNotification := func() error {
		waitCancel()

		// проверяется, действительно ошибка связана с необходимостью следующей
		// проверки соединения, если нет, то значит, что-то не так с самим соединением
		if err := <-errChan; err != nil {
			if !errors.Is(err, context.Canceled) {
				return fmt.Errorf("interruptWaitForNotification: %w", err)
			}
		}

		return nil
	}

	timeForCheckCtx, timeForCheckCancel := context.WithTimeout(ctx, p.checkConnPeriod)
	defer timeForCheckCancel()

	// Ожидается одно из следующих событий:
	//   - срабатывание таймаута для проверки активности соединения;
	//   - ошибка соединения для WaitForNotification;
	//   - завершение основного процесса сервиса (Shutdown);
	select {
	case <-timeForCheckCtx.Done():
		if err := interruptWaitForNotification(); err != nil {
			return err
		}

		if err := conn.Ping(ctx); err != nil {
			return fmt.Errorf("waitSomePeriod.conn.Ping: %w", err)
		}

		p.logger.Debug(ctx, "conn.Ping is successful")
	case err := <-errChan:
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				return fmt.Errorf("waitSomePeriod: %w", err)
			}
		}

		p.logger.Debug(ctx, "Received canceled event from errChan")
	case <-ctx.Done():
		return <-errChan // дожидается остановки WaitForNotification
	}

	return nil
}

// Find - находит канал по имени и возвращает его для получения уведомлений.
func (p *ProcessWaitForNotification) Find(name string) (<-chan struct{}, error) {
	return p.receiverChannels.Find(name)
}

// MustFind - находит канал по имени и возвращает его для получения уведомлений.
// Если имя канала не зарегистрировано, то регистрируется ошибка и возвращается
// фиктивный канал, который будет заблокирован до завершения процесса.
func (p *ProcessWaitForNotification) MustFind(name string) <-chan struct{} {
	ch, err := p.receiverChannels.Find(name)
	if err != nil {
		mrlog.Error(p.logger, "ProcessWaitForNotification.MustFind", "error", err)

		return p.done
	}

	return ch
}

// Shutdown - корректно завершает процесс прослушивания NOTIFY.
func (p *ProcessWaitForNotification) Shutdown(ctx context.Context) error {
	p.logger.Info(ctx, "Shutting down the WaitForNotification...")
	close(p.done)

	p.wg.Wait()
	p.logger.Info(ctx, "The WaitForNotification has been shut down")

	return nil
}
