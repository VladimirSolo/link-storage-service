package postgres

import (
	"context"
	"database/sql"
	"link-storage-service/internal/domain"

	"github.com/lib/pq"
)

type LinkRepository struct {
	db *sql.DB
}

func NewLinkRepository(db *sql.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) Create(ctx context.Context, link *domain.Link) error {
	query := `
		INSERT INTO links (short_code, original_url, created_at, visits)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	err := r.db.QueryRowContext(ctx, query,
		link.ShortCode, link.OriginalURL, link.CreatedAt, link.Visits,
	).Scan(&link.ID)
	if err != nil && isUniqueViolation(err) {
		return domain.ErrConflict
	}
	return err
}

func (r *LinkRepository) GetByShortCode(ctx context.Context, shortCode string) (*domain.Link, error) {
	link := &domain.Link{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, short_code, original_url, created_at, visits FROM links WHERE short_code = $1`,
		shortCode,
	).Scan(&link.ID, &link.ShortCode, &link.OriginalURL, &link.CreatedAt, &link.Visits)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (r *LinkRepository) List(ctx context.Context, limit, offset int) ([]*domain.Link, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, short_code, original_url, created_at, visits FROM links ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*domain.Link
	for rows.Next() {
		l := &domain.Link{}
		if err := rows.Scan(&l.ID, &l.ShortCode, &l.OriginalURL, &l.CreatedAt, &l.Visits); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (r *LinkRepository) Delete(ctx context.Context, shortCode string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM links WHERE short_code = $1`, shortCode)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *LinkRepository) IncrementVisits(ctx context.Context, shortCode string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE links SET visits = visits + 1 WHERE short_code = $1`,
		shortCode,
	)
	return err
}

func isUniqueViolation(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code == "23505"
}
