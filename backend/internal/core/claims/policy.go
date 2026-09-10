package claims

import (
	"context"
	"errors"
	"github.com/casbin/casbin/v3"
	"gorm.io/gorm"
	"sync"
)

type PolicyRevision struct {
	ID      uint   `gorm:"primaryKey"`
	Version uint64 `gorm:"not null"`
}

func (PolicyRevision) TableName() string { return "sys_policy_revision" }

// PolicyTransaction serializes policy writers and publishes a revision in the
// same commit as the rules. A failed transaction cannot publish a new revision.
func PolicyTransaction(ctx context.Context, db *gorm.DB, update func(*gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&PolicyRevision{}).Where("id = 1").Update("version", gorm.Expr("version + 1"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errors.New("policy schema not initialized; run migrate")
		}
		return update(tx)
	})
}

type PolicyManager struct {
	db       *gorm.DB
	enforcer *casbin.SyncedCachedEnforcer
	mu       sync.Mutex
	revision uint64
}

func NewPolicyManager(db *gorm.DB) (*PolicyManager, error) {
	enforcer, err := InitCasbin(db)
	if err != nil {
		return nil, err
	}
	return &PolicyManager{db: db, enforcer: enforcer}, nil
}

func (p *PolicyManager) Enforce(ctx context.Context, subject, path, method string) (bool, error) {
	// Every instance reads the committed version. Do not cache this DB read:
	// revocations must be effective on the first request after the write returns.
	p.mu.Lock()
	defer p.mu.Unlock()
	var revision PolicyRevision
	if err := p.db.WithContext(ctx).First(&revision, 1).Error; err != nil {
		return false, err
	}
	if revision.Version != p.revision {
		if err := p.enforcer.LoadPolicy(); err != nil {
			return false, err
		}
		p.revision = revision.Version
	}
	return p.enforcer.Enforce(subject, path, method)
}
