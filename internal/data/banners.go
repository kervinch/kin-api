package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kervinch/internal/validator"
)

type Banner struct {
	ID          int64     `json:"id"`
	ImageURL    string    `json:"image_url"`
	Title       string    `json:"title"`
	Deeplink    string    `json:"deeplink"`
	OutboundURL string    `json:"outbound_url"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

func ValidateBanner(v *validator.Validator, banner *Banner) {
	v.Check(banner.ImageURL != "", "image_url", "must be provided")
	v.Check(banner.Title != "", "title", "must be provided")
	v.Check(len(banner.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(len(banner.Deeplink) <= 500, "deeplink", "must not be more than 500 bytes long")
	v.Check(len(banner.OutboundURL) <= 500, "outbound_url", "must not be more than 500 bytes long")
}

type BannerModel struct {
	DB *sql.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m BannerModel) Insert(banner *Banner) error {
	query := `
		INSERT INTO banners (image_url, title, deeplink, outbound_url, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, image_url, title`

	args := []interface{}{banner.ImageURL, banner.Title, banner.Deeplink, banner.OutboundURL, banner.IsActive}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&banner.ID, &banner.ImageURL, &banner.Title)
}

func (m BannerModel) GetAll(title string, filters Filters) ([]*Banner, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, image_url, title, deeplink, outbound_url
		FROM banners
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (is_active = true)
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{title, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	banners := []*Banner{}

	for rows.Next() {
		var banner Banner

		err := rows.Scan(
			&totalRecords,
			&banner.ID,
			&banner.ImageURL,
			&banner.Title,
			&banner.Deeplink,
			&banner.OutboundURL,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		banners = append(banners, &banner)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return banners, metadata, nil
}

func (m BannerModel) Get(id int64) (*Banner, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, image_url, title, deeplink, outbound_url, is_active
		FROM banners
		WHERE id = $1`

	var banner Banner

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&banner.ID,
		&banner.ImageURL,
		&banner.Title,
		&banner.Deeplink,
		&banner.OutboundURL,
		&banner.IsActive,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &banner, nil
}

func (m BannerModel) Update(banner *Banner) error {
	query := `
		UPDATE banners
		SET image_url = $1, title = $2, deeplink = $3, outbound_url = $4, is_active = $5
		WHERE id = $6
		RETURNING id, image_url, title, deeplink, outbound_url, is_active`

	args := []interface{}{
		banner.ImageURL,
		banner.Title,
		banner.Deeplink,
		banner.OutboundURL,
		banner.IsActive,
		banner.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&banner.ID, &banner.ImageURL, &banner.Title, &banner.Deeplink, &banner.OutboundURL, &banner.IsActive)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m BannerModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM banners
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m BannerModel) GetAPI() ([]*Banner, error) {
	query := fmt.Sprintln(`
		SELECT id, image_url, title, deeplink, outbound_url, is_active
		FROM banners
		WHERE (is_active = true)
		ORDER BY id ASC
		LIMIT 5`)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	banners := []*Banner{}

	for rows.Next() {
		var banner Banner

		err := rows.Scan(
			&banner.ID,
			&banner.ImageURL,
			&banner.Title,
			&banner.Deeplink,
			&banner.OutboundURL,
			&banner.IsActive,
		)
		if err != nil {
			return nil, err
		}

		banners = append(banners, &banner)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return banners, nil
}
