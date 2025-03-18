// internal/storage/storage.go
package storage

import (
	"context"
	"errors"
	"songlibrary/internal/models"
)

var ErrSongNotFound = errors.New("song not found")

type SongStorage interface {
	Create(ctx context.Context, song *models.Song) (*models.Song, error)
	GetByID(ctx context.Context, id int) (*models.Song, error)
	List(ctx context.Context, filter *models.SongFilter, pagination *models.Pagination) ([]models.Song, error)
	Update(ctx context.Context, song *models.Song) (*models.Song, error)
	Delete(ctx context.Context, id int) error
}
