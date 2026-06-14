package service

import (
	"context"
	"net/url"
	"shortify/server/internal/domain"
	"shortify/server/internal/repository"
	"shortify/server/internal/worker"
	"shortify/server/pkg/shortcode"
	"strings"

	"github.com/google/uuid"
)

type LinkService struct {
	links       *repository.LinkRepository
	clicks      *repository.ClickRepository
	clickWorker *worker.ClickWorker
	baseURL     string
}

func NewLinkService(
	links *repository.LinkRepository,
	clicks *repository.ClickRepository,
	clickWorker *worker.ClickWorker,
	baseURL string,
) *LinkService {
	return &LinkService{
		links:       links,
		clicks:      clicks,
		clickWorker: clickWorker,
		baseURL:     strings.TrimRight(baseURL, "/"),
	}
}

type LinkView struct {
	ID          uuid.UUID `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	ShortURL    string    `json:"short_url"`
	ClicksCount int64     `json:"clicks_count"`
	CreatedAt   string    `json:"created_at"`
}

func (s *LinkService) Create(ctx context.Context, userID uuid.UUID, originalURL string) (*LinkView, error) {
	originalURL = strings.TrimSpace(originalURL)
	parsed, err := url.ParseRequestURI(originalURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, domain.ErrInvalidInput
	}

	code, err := shortcode.Generate(8)
	if err != nil {
		return nil, err
	}

	link := &domain.Link{
		ID:          uuid.New(),
		UserID:      userID,
		OriginalURL: originalURL,
		ShortCode:   code,
	}

	if err := s.links.Create(ctx, link); err != nil {
		return nil, err
	}

	return s.toView(ctx, link)
}

func (s *LinkService) List(ctx context.Context, userID uuid.UUID) ([]LinkView, error) {
	links, err := s.links.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]LinkView, 0, len(links))
	for i := range links {
		view, err := s.toView(ctx, &links[i])
		if err != nil {
			return nil, err
		}
		result = append(result, *view)
	}

	return result, nil
}

func (s *LinkService) Delete(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) error {
	return s.links.Delete(ctx, userID, linkID)
}

func (s *LinkService) Resolve(ctx context.Context, code string, remoteAddr string, userAgent string) (string, error) {
	link, err := s.links.FindByShortCode(ctx, code)
	if err != nil {
		return "", err
	}

	s.clickWorker.Enqueue(worker.ClickEvent{
		LinkID:    link.ID,
		IP:        remoteAddr,
		UserAgent: userAgent,
	})

	return link.OriginalURL, nil
}

func (s *LinkService) GetStats(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) (int64, []domain.Click, error) {
	link, err := s.links.FindByID(ctx, linkID)
	if err != nil {
		return 0, nil, err
	}
	if link.UserID != userID {
		return 0, nil, domain.ErrForbidden
	}

	total, err := s.clicks.CountByLink(ctx, linkID)
	if err != nil {
		return 0, nil, err
	}

	recent, err := s.clicks.ListByLink(ctx, linkID, 20)
	if err != nil {
		return 0, nil, err
	}

	return total, recent, nil
}

func (s *LinkService) toView(ctx context.Context, link *domain.Link) (*LinkView, error) {
	count, err := s.clicks.CountByLink(ctx, link.ID)
	if err != nil {
		return nil, err
	}

	return &LinkView{
		ID:          link.ID,
		OriginalURL: link.OriginalURL,
		ShortCode:   link.ShortCode,
		ShortURL:    s.baseURL + "/" + link.ShortCode,
		ClicksCount: count,
		CreatedAt:   link.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
