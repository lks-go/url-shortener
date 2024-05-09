package urldeleter

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/lks-go/url-shortener/internal/service"
)

type Config struct {
	StoppingTimeout  time.Duration
	MaxBatchSize     int
	BatchWaitingTime time.Duration
}

type Deps struct {
	Storage service.URLStorage
}

func NewDeleter(cfg Config, d Deps) *URLDeleter {

	if cfg.StoppingTimeout <= 0 {
		cfg.StoppingTimeout = time.Second * 1
	}

	if cfg.MaxBatchSize == 0 {
		cfg.MaxBatchSize = 10
	}

	if cfg.BatchWaitingTime == 0 {
		cfg.BatchWaitingTime = time.Millisecond * 100
	}

	return &URLDeleter{
		cfg:     cfg,
		storage: d.Storage,
		queue:   make(chan string),
		stopped: make(chan struct{}),
	}
}

type URLDeleter struct {
	cfg     Config
	storage service.URLStorage
	queue   chan string
	stopped chan struct{}
}

func (d *URLDeleter) Start() {
	listToDelete := make([]string, 0, d.cfg.MaxBatchSize)
	send, sendAndExit := false, false

	go func() {
		ticker := time.NewTicker(d.cfg.BatchWaitingTime)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if len(listToDelete) == 0 {
					continue
				}
				send = true

			case v, ok := <-d.queue:
				if !ok {
					sendAndExit = true
					break
				}

				listToDelete = append(listToDelete, v)
				if len(listToDelete) == d.cfg.MaxBatchSize {
					send = true
				}
			}

			if send || sendAndExit {
				send = false

				if err := d.storage.DeleteURLs(context.Background(), listToDelete); err != nil {
					logrus.Errorf("filed to delete urls: %s", err)
				}

				listToDelete = make([]string, 0, d.cfg.MaxBatchSize)
			}

			if sendAndExit {
				return
			}
		}
	}()
}

func (d *URLDeleter) Stop() {
	close(d.stopped)
	time.Sleep(d.cfg.StoppingTimeout)
	close(d.queue)
}

func (d *URLDeleter) Delete(ctx context.Context, userID string, codes []string) error {
	select {
	case <-d.stopped:
		return service.ErrURLDeleterStopped
	default:
	}

	belongCodes, err := d.storage.UsersURLCodes(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user codes: %w", err)
	}

	for _, code := range codes {
		if isBelong(belongCodes, code) {
			d.queue <- code
		}
	}

	return nil
}

func isBelong(belongCodes []string, code string) bool {
	for _, c := range belongCodes {
		if c == code {
			return true
		}
	}

	return false
}
