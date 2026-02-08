package database

import (
	"context"
	"fmt"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
)

const (
	ReversalsCollectionName = "reversals"
)

type ReversalRepository struct {
	collection Collection
}

func NewReversalRepository() *ReversalRepository {
	return &ReversalRepository{collection: NewMongoCollection(GetCollection(DBName, ReversalsCollectionName))}
}

// NewReversalRepositoryWithCollection creates a reversal repository with custom collection (for testing)
func NewReversalRepositoryWithCollection(c Collection) *ReversalRepository {
	return &ReversalRepository{collection: c}
}

func (rr *ReversalRepository) CreateReversalRecord(ctx context.Context, r *models.ReversalRecord) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	r.CreatedAt = time.Now()
	_, err := rr.collection.InsertOne(ctx, r)
	if err != nil {
		return fmt.Errorf("failed to create reversal record: %w", err)
	}
	return nil
}
