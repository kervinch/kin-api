package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type InboxUser struct {
	ID        int64     `json:"id"`
	Inbox     Inbox     `json:"inbox"`
	InboxID   int64     `json:"inbox_id"`
	UserID    int64     `json:"user_id"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func ValidateInboxUsers(v *validator.Validator, inboxUser *InboxUser) {
	v.Check(inboxUser.InboxID != 0, "inbox_id", "must be provided")
	v.Check(inboxUser.InboxID > 0, "inbox_id", "must be a positive integer")
	v.Check(inboxUser.UserID != 0, "user_id", "must be provided")
	v.Check(inboxUser.UserID > 0, "user_id", "must be a positive integer")
}

type InboxUserModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m InboxUserModel) Insert(inboxUser *InboxUser) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Create(&inboxUser).Error

	return err
}

func (m InboxUserModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("inbox_id = ?", id).Delete(&InboxUser{}, id).Error
	if err != nil {
		return err
	}

	return nil
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m InboxUserModel) Get(inboxID int64, user *User) (*InboxUser, error) {
	var inboxUser *InboxUser

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("id = ? AND user_id = ?", inboxID, user.ID).First(&inboxUser).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return inboxUser, nil
}
