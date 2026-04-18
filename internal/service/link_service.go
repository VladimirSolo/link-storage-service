package service

import (
	"context"
	"crypto/rand"
	"errors"
	"link-storage-service/internal/domain"
	"time"
)

const (
	shortCodeLen = 6
	charset      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxRetries   = 5
)

//go:generate mockgen -destination=mock_service.go -package=service . LinkRepository,LinkCache

type LinkRepository interface {
	Create(ctx context.Context, link *domain.Link) error
	GetByShortCode(ctx context.Context, shortCode string) (*domain.Link, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Link, error)
	Delete(ctx context.Context, shortCode string) error
	IncrementVisits(ctx context.Context, shortCode string) error
}

type LinkCache interface {
	Get(shortCode string) (*domain.Link, bool)
	Set(shortCode string, link *domain.Link)
	Delete(shortCode string)
}

type LinkService struct {
	repo  LinkRepository
	cache LinkCache
}

func NewLinkService(repo LinkRepository, cache LinkCache) *LinkService {
	return &LinkService{repo: repo, cache: cache}
}

func (s *LinkService) Create(ctx context.Context, originalURL string) (*domain.Link, error) {
	for i := 0; i < maxRetries; i++ {
		code, err := generateShortCode()
		if err != nil {
			return nil, err
		}
		link := &domain.Link{
			ShortCode:   code,
			OriginalURL: originalURL,
			CreatedAt:   time.Now().UTC(),
		}
		err = s.repo.Create(ctx, link)
		if err == nil {
			return link, nil
		}
		if !errors.Is(err, domain.ErrConflict) {
			return nil, err
		}
		// collision — retry with a new code
	}
	return nil, errors.New("failed to generate unique short code after retries")
}

// Get returns the link and increments the visit counter.
func (s *LinkService) Get(ctx context.Context, shortCode string) (*domain.Link, error) {
	if cached, ok := s.cache.Get(shortCode); ok {
		if err := s.repo.IncrementVisits(ctx, shortCode); err != nil {
			return nil, err
		}
		// Return a copy with incremented visits to reflect this request.
		result := *cached
		result.Visits++
		s.cache.Delete(shortCode) // invalidate so next read reflects accurate DB state
		return &result, nil
	}

	link, err := s.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	if err := s.repo.IncrementVisits(ctx, shortCode); err != nil {
		return nil, err
	}
	link.Visits++
	s.cache.Set(shortCode, link)
	return link, nil
}

func (s *LinkService) List(ctx context.Context, limit, offset int) ([]*domain.Link, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.List(ctx, limit, offset)
}

func (s *LinkService) Delete(ctx context.Context, shortCode string) error {
	if err := s.repo.Delete(ctx, shortCode); err != nil {
		return err
	}
	s.cache.Delete(shortCode)
	return nil
}

// Stats returns link info without incrementing visits.
func (s *LinkService) Stats(ctx context.Context, shortCode string) (*domain.Link, error) {
	return s.repo.GetByShortCode(ctx, shortCode)
}

func generateShortCode() (string, error) {
	b := make([]byte, shortCodeLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i, v := range b {
		b[i] = charset[int(v)%len(charset)]
	}
	return string(b), nil
}
