package database

import (
	"context"
	"testing"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReversalRepository_CreateReversalRecord_Mock(t *testing.T) {
	mockCollection := NewMockCollection()
	mockCollection.On("InsertOne", mock.Anything, mock.MatchedBy(func(doc interface{}) bool {
		rev, ok := doc.(*models.ReversalRecord)
		return ok && rev.ID == "rev-1"
	})).Return(&mongo.InsertOneResult{InsertedID: "rev-1"}, nil)

	repo := NewReversalRepositoryWithCollection(mockCollection)
	ctx := context.Background()

	rev := &models.ReversalRecord{
		ID:        "rev-1",
		InvoiceID: "inv-1",
		Amount:    100.0,
		Date:      "2026-02-08",
		Phone:     "254700000001",
		AdminID:   "admin-1",
		Reason:    "Customer refund",
	}

	err := repo.CreateReversalRecord(ctx, rev)

	assert.NoError(t, err)
	mockCollection.AssertCalled(t, "InsertOne", mock.Anything, mock.Anything)
	assert.NotZero(t, rev.CreatedAt)
}

func TestReversalRepository_CreateReversalRecord_Error_Mock(t *testing.T) {
	mockCollection := NewMockCollection()
	mockCollection.On("InsertOne", mock.Anything, mock.Anything).
		Return(nil, mongo.ErrNilDocument)

	repo := NewReversalRepositoryWithCollection(mockCollection)
	ctx := context.Background()

	rev := &models.ReversalRecord{
		ID:        "rev-bad",
		InvoiceID: "inv-1",
		Amount:    50.0,
		Reason:    "Test error",
	}

	err := repo.CreateReversalRecord(ctx, rev)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create reversal record")
}
