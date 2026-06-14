package worker

import (
	"context"
	"log/slog"
	"shortify/server/internal/domain"
	"shortify/server/internal/repository"
	"sync"

	"github.com/google/uuid"
)

type ClickEvent struct {
	LinkID    uuid.UUID
	IP        string
	UserAgent string
}

type ClickWorker struct {
	clicks *repository.ClickRepository
	events chan ClickEvent
	wg     sync.WaitGroup
}

func NewClickWorker(clicks *repository.ClickRepository, buffer int) *ClickWorker {
	return &ClickWorker{
		clicks: clicks,
		events: make(chan ClickEvent, buffer),
	}
}

func (w *ClickWorker) Start(ctx context.Context) {
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()

		for {
			select {
			case <-ctx.Done():
				w.drain(context.Background())
				return
			case event := <-w.events:
				w.save(ctx, event)
			}
		}
	}()
}

func (w *ClickWorker) drain(ctx context.Context) {
	for {
		select {
		case event := <-w.events:
			w.save(ctx, event)
		default:
			return
		}
	}
}

func (w *ClickWorker) Enqueue(event ClickEvent) {
	select {
	case w.events <- event:
	default:
		slog.Warn("click event dropped")
	}
}

func (w *ClickWorker) save(ctx context.Context, event ClickEvent) {
	click := &domain.Click{
		ID:        uuid.New(),
		LinkID:    event.LinkID,
		IP:        event.IP,
		UserAgent: event.UserAgent,
	}

	err := w.clicks.Create(ctx, click)
	if err != nil {
		slog.Error("failed to save click", "error", err)
	}
}
