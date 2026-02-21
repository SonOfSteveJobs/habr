package closer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/SonOfSteveJobs/habr/pkg/logger"
)

const shutdownTimeout = 5 * time.Second

type Closer struct {
	mu    sync.Mutex
	once  sync.Once
	done  chan struct{}
	funcs []func(context.Context) error
}

var globalCloser = NewCloser()

func NewCloser() *Closer {
	return &Closer{
		done: make(chan struct{}),
	}
}

// Add добавляет функции закрытия
func Add(f ...func(context.Context) error) {
	globalCloser.Add(f...)
}

// AddNamed добавляет функцию закрытия с именем ресурса для логирования
func AddNamed(name string, f func(context.Context) error) {
	globalCloser.AddNamed(name, f)
}

// CloseAll вызывает все зарегистрированные функции закрытия
func CloseAll(ctx context.Context) error {
	return globalCloser.CloseAll(ctx)
}

// Wait ожидает завершения graceful shutdown
func Wait() {
	<-globalCloser.done
}

// Listen слушает системные сигналы и запускает graceful shutdown
func Listen(signals ...os.Signal) {
	go globalCloser.listen(signals...)
}

func (c *Closer) Add(f ...func(context.Context) error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.funcs = append(c.funcs, f...)
}

func (c *Closer) AddNamed(name string, f func(context.Context) error) {
	c.Add(func(ctx context.Context) error {
		start := time.Now()
		log := logger.Logger()

		log.Info().Str("resource", name).Msg("closing resource")

		err := f(ctx)
		duration := time.Since(start)

		if err != nil {
			log.Error().Err(err).Str("resource", name).Dur("duration", duration).Msg("failed to close resource")
		} else {
			log.Info().Str("resource", name).Dur("duration", duration).Msg("resource closed")
		}

		return err
	})
}

func (c *Closer) CloseAll(ctx context.Context) error {
	var result error

	c.once.Do(func() {
		defer close(c.done)

		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		if len(funcs) == 0 {
			return
		}

		log := logger.Logger()
		log.Info().Msg("starting graceful shutdown")

		errCh := make(chan error, len(funcs))
		var wg sync.WaitGroup

		for i := len(funcs) - 1; i >= 0; i-- {
			f := funcs[i]

			wg.Go(func() {
				defer func() {
					if r := recover(); r != nil {
						errCh <- fmt.Errorf("panic in closer: %v", r)
					}
				}()

				if err := f(ctx); err != nil {
					errCh <- err
				}
			})
		}

		go func() {
			wg.Wait()
			close(errCh)
		}()

		for {
			select {
			case <-ctx.Done():
				log.Error().Err(ctx.Err()).Msg("shutdown timeout exceeded")

				if result == nil {
					result = ctx.Err()
				}

				return
			case err, ok := <-errCh:
				if !ok {
					log.Info().Msg("graceful shutdown completed")
					return
				}

				log.Error().Err(err).Msg("shutdown error")

				if result == nil {
					result = err
				}
			}
		}
	})

	return result
}

func (c *Closer) listen(signals ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	defer signal.Stop(ch)

	select {
	case sig := <-ch:
		log := logger.Logger()
		log.Info().Str("signal", sig.String()).Msg("received signal, shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := c.CloseAll(ctx); err != nil {
			log.Error().Err(err).Msg("shutdown finished with errors")
		}

		go func() {
			ch2 := make(chan os.Signal, 1)
			signal.Notify(ch2, signals...)
			<-ch2
			log.Warn().Msg("forced exit")
			os.Exit(1)
		}()

	case <-c.done:
		return
	}

	<-c.done
}
