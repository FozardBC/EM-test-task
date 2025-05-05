package storage

import (
	"context"
	"errors"
	"test-task/internal/domain/filters"
	"test-task/internal/domain/models"
)

var (
	ErrIDNotFound = errors.New("ID not found")
)

type Storage interface {
	Save(ctx context.Context, entity *models.Person) (int64, error)
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*models.Person, error)
	Update(ctx context.Context, entity *models.Person, id int64) error
	FilteredPages(ctx context.Context, offset int, limit int, options *filters.Options) ([]*models.Person, int, error)
	Close()
	Ping(ctx context.Context) error
}
