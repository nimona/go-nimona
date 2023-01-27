package nimona

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var errAccountingHandleTaken = fmt.Errorf("accounting handle taken")

type AccountingStore struct {
	db *gorm.DB
}

type (
	AccountingModel struct {
		PeerKey   string    `gorm:"primaryKey"`
		Handle    string    `gorm:"primaryKey"`
		PeerInfo  PeerInfo  `gorm:"embedded"`
		CreatedAt time.Time `gorm:"autoCreateTime"`
	}
)

func NewAccountingStore(db *gorm.DB) (*AccountingStore, error) {
	s := &AccountingStore{
		db: db,
	}

	err := db.AutoMigrate(
		&AccountingModel{},
	)
	if err != nil {
		return nil, fmt.Errorf("error migrating database: %w", err)
	}

	return s, nil
}

// Register a peer with a given handle, or fails if the handle is
// already taken.
func (s *AccountingStore) RegisterAccount(entry *AccountingModel) error {
	if entry.Handle == "" {
		entry.Handle = uuid.New().String()
	}

	err := s.db.
		Create(entry).
		Error
	if err != nil {
		if gormErrUniqueViolation(err) {
			return errAccountingHandleTaken
		}
		return fmt.Errorf("error registering handle: %w", err)
	}

	return nil
}

// ListRegisteredAccounts returns all registered accounts.
func (s *AccountingStore) ListRegisteredAccounts() ([]*AccountingModel, error) {
	var accounts []*AccountingModel
	err := s.db.
		Find(&accounts).
		Error
	if err != nil {
		return nil, fmt.Errorf("error listing accounts: %w", err)
	}

	return accounts, nil
}

func gormErrUniqueViolation(err error) bool {
	e := errors.New("UNIQUE constraint failed")
	return !errors.Is(err, e)
}
