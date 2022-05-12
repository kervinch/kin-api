package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type Inbox struct {
	ID        int64       `json:"id"`
	Title     string      `json:"title"`
	Content   string      `json:"content"`
	ImageURL  string      `json:"image_url"`
	Deeplink  string      `json:"deeplink"`
	Slug      string      `json:"slug"`
	InboxUser []InboxUser `json:"inbox_users"`
	CreatedAt time.Time   `json:"-"`
	UpdatedAt time.Time   `json:"-"`
}

func ValidateInbox(v *validator.Validator, inbox *Inbox) {
	v.Check(inbox.Title != "", "title", "must be provided")
	v.Check(inbox.Content != "", "content", "must be provided")
	v.Check(inbox.Slug != "", "slug", "must be provided")
}

type InboxModel struct {
	DB *gorm.DB
}

func (Inbox) TableName() string {
	return "inbox"
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m InboxModel) GetAll(p Pagination) ([]*Inbox, Metadata, error) {
	var inbox []*Inbox
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Order("created_at DESC").Find(&inbox).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("inbox").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return inbox, metadata, nil
}

func (m InboxModel) Get(id int64) (*Inbox, error) {
	var inbox *Inbox

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("id = ?", id).First(&inbox).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return inbox, nil
}

func (m InboxModel) Insert(inbox *Inbox) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Create(&inbox).Error
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "inbox_slug_key"`:
			return ErrDuplicateSlug
		default:
			return err
		}
	}

	return err
}

func (m InboxModel) Update(i *Inbox) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var inbox *Inbox

	err := m.DB.WithContext(ctx).First(&inbox, i.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	inbox.Title = i.Title
	inbox.Content = i.Content
	inbox.ImageURL = i.ImageURL
	inbox.Deeplink = i.Deeplink
	inbox.Slug = i.Slug

	err = m.DB.Save(&inbox).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrEditConflict
		case err.Error() == `pq: duplicate key value violates unique constraint "inbox_slug_key"`:
			return ErrDuplicateSlug
		default:
			return err
		}
	}

	return nil
}

func (m InboxModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ra := m.DB.WithContext(ctx).Where("id = ?", id).Delete(&Inbox{}).RowsAffected
	if ra < 1 {
		return ErrRecordNotFound
	}

	return nil
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m InboxModel) GetAPI(p Pagination, user *User) ([]*Inbox, Metadata, error) {
	var inbox []*Inbox
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Preload("InboxUser").Order("created_at DESC").Find(&inbox).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("inbox").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return inbox, metadata, nil
}

func (m InboxModel) GetBySlug(slug string) (*Inbox, error) {
	if slug == "" {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var inbox *Inbox

	err := m.DB.WithContext(ctx).Where("slug = ?", slug).First(&inbox).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return inbox, nil
}
