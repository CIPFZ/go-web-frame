package session

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

// Record persists one login family. JWT refresh retains ID, so logout revokes
// both the original token and every refreshed token from that login.
type Record struct {
	ID           string `gorm:"primaryKey;size:36"`
	UserID       uint   `gorm:"index;not null"`
	TokenVersion uint64 `gorm:"not null"`
	CreatedAt    time.Time
	ExpiresAt    time.Time `gorm:"index;not null"`
	RevokedAt    *time.Time
}

func (Record) TableName() string { return "sys_user_sessions" }

type Identity struct {
	ID          string
	UserID      uint
	Version     uint64
	AuthorityID uint
}

var ErrInvalid = errors.New("session revoked or expired")

type Store struct{ db *gorm.DB }

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

func (s *Store) Create(ctx context.Context, record Record) error {
	return s.db.WithContext(ctx).Create(&record).Error
}

func (s *Store) Validate(ctx context.Context, identity Identity) error {
	if identity.ID == "" {
		return ErrInvalid
	}
	var count int64
	err := s.db.WithContext(ctx).Table("sys_user_sessions AS s").
		Joins("JOIN sys_users AS u ON u.id = s.user_id").
		Joins("JOIN sys_authorities AS a ON a.authority_id = u.authority_id").
		Joins("JOIN sys_user_authorities AS ua ON ua.user_id = u.id AND ua.authority_id = u.authority_id").
		Where("s.id = ? AND s.user_id = ? AND s.token_version = ? AND s.revoked_at IS NULL AND s.expires_at > ?", identity.ID, identity.UserID, identity.Version, time.Now()).
		Where("u.token_version = ? AND u.status = 1 AND u.deleted_at IS NULL AND u.authority_id = ? AND a.deleted_at IS NULL", identity.Version, identity.AuthorityID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrInvalid
	}
	return nil
}

func (s *Store) Renew(ctx context.Context, identity Identity, expires time.Time) error {
	if err := s.Validate(ctx, identity); err != nil {
		return err
	}
	result := s.db.WithContext(ctx).Model(&Record{}).Where("id = ? AND revoked_at IS NULL AND expires_at < ?", identity.ID, expires).Update("expires_at", expires)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// Concurrent refreshes can propose the same expiry; MySQL reports no
		// changed rows. A still-valid family is not a failed refresh.
		return s.Validate(ctx, identity)
	}
	return nil
}
func (s *Store) Revoke(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Model(&Record{}).Where("id = ? AND revoked_at IS NULL", id).Update("revoked_at", time.Now()).Error
}
func (s *Store) Cleanup(ctx context.Context) error {
	return s.db.WithContext(ctx).Where("expires_at < ?", time.Now().Add(-24*time.Hour)).Delete(&Record{}).Error
}
