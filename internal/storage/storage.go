package storage

import (
	"context"
	"database/sql"
	"errors"
	"songlibrary/internal/models"
)

var ErrSongNotFound = errors.New("song not found")

//go:generate mockgen -destination=mocks/mock_storage.go -package=mocks songlibrary/internal/storage SongStorage

type SongStorage interface {
	Create(ctx context.Context, song *models.Song, tx *sql.Tx) (*models.Song, error)
	GetByID(ctx context.Context, id int) (*models.Song, error)
	List(ctx context.Context, filter *models.SongFilter, pagination *models.Pagination) ([]models.Song, error)
	Update(ctx context.Context, song *models.Song) (*models.Song, error)
	Delete(ctx context.Context, id int) error
	BeginTx(ctx context.Context) (*sql.Tx, error)
}
