package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type Blog struct {
	ID             int64        `json:"id"`
	BlogCategory   BlogCategory `json:"blog_category"`
	BlogCategoryID int          `json:"blog_category_id"`
	Thumbnail      string       `json:"thumbnail"`
	Title          string       `json:"title"`
	Description    string       `json:"description"`
	Content        string       `json:"content"`
	Slug           string       `json:"slug"`
	Type           string       `json:"type"`
	PublishedAt    time.Time    `json:"published_at"`
	Feature        bool         `json:"feature"`
	Status         string       `json:"status"`
	Tags           string       `json:"tags"`
	CreatedBy      int          `json:"created_by"`
	DeletedAt      time.Time    `json:"-"`
	CreatedAt      time.Time    `json:"-"`
	UpdatedAt      time.Time    `json:"-"`
	CreatedByText  string       `json:"created_by_text"`
}

func ValidateBlog(v *validator.Validator, blog *Blog) {
	v.Check(blog.Thumbnail != "", "thumbnail", "must be provided")
	v.Check(blog.Title != "", "title", "must be provided")
	v.Check(len(blog.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(blog.Description != "", "description", "must be provided")
	v.Check(blog.Content != "", "content", "must be provided")
}

type BlogModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m BlogModel) GetAll(p Pagination) ([]*Blog, Metadata, error) {
	var blogs []*Blog
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Order("status").Preload("BlogCategory").Find(&blogs).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("blogs").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return blogs, metadata, nil
}

func (m BlogModel) Get(id int64) (*Blog, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	var blog *Blog

	err := m.DB.Preload("BlogCategory").First(&blog, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return blog, nil
}

func (m BlogModel) Insert(blog *Blog) error {
	err := m.DB.Create(&blog).Error
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "blogs_slug_key"`:
			return ErrDuplicateSlug
		default:
			return err
		}
	}

	return err
}

func (m BlogModel) Update(b *Blog) error {
	var blog *Blog

	err := m.DB.First(&blog, b.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	blog.BlogCategoryID = b.BlogCategoryID
	blog.Thumbnail = b.Thumbnail
	blog.Title = b.Title
	blog.Description = b.Description
	blog.Content = b.Content
	blog.Slug = b.Slug
	blog.Type = b.Type
	blog.Feature = b.Feature
	blog.Status = b.Status
	blog.Tags = b.Tags

	err = m.DB.Save(&blog).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrEditConflict
		case err.Error() == `pq: duplicate key value violates unique constraint "blogs_slug_key"`:
			return ErrDuplicateSlug
		default:
			return err
		}
	}

	return nil
}

func (m BlogModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ra := m.DB.Delete(&Blog{}, id).RowsAffected
	if ra < 1 {
		return ErrRecordNotFound
	}

	return nil
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m BlogModel) GetAPI(p Pagination) ([]*Blog, Metadata, error) {
	var blogs []*Blog
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Where("status = ?", "published").Preload("BlogCategory").Find(&blogs).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("blogs").Where("status = ?", "published").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return blogs, metadata, nil
}

func (m BlogModel) GetBySlug(slug string) (*Blog, error) {
	if slug == "" {
		return nil, ErrRecordNotFound
	}

	var blog *Blog

	err := m.DB.Preload("BlogCategory").Where("slug = ?", slug).First(&blog).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return blog, nil
}

func (m BlogModel) GetRecommendations() ([]*Blog, error) {
	var blogs []*Blog

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("status = ?", "published").Order("created_at desc").Limit(8).Preload("BlogCategory").Find(&blogs).Error
	if err != nil {
		return nil, err
	}

	return blogs, nil
}
