package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"songlibrary/internal/lib/logger/utils"
	"songlibrary/internal/models"
	"songlibrary/internal/storage"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type PgStorage struct {
	conn *pgx.Conn
}

func NewPgStorage(conn *pgx.Conn) storage.SongStorage {
	return &PgStorage{conn: conn}
}

// BeginTx starts a new transaction.
func (s *PgStorage) BeginTx(ctx context.Context) (*sql.Tx, error) {
	db, err := sql.Open("pgx", s.conn.Config().ConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to open sql connection: %w", err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	return tx, nil
}

// Create creates a new song.
func (s *PgStorage) Create(ctx context.Context, song *models.Song, tx *sql.Tx) (*models.Song, error) {
	var query string
	var addedSong models.Song
	var err error

	if tx != nil {
		query = `
            INSERT INTO songs (group_name, song_name, release_date, text, link)
            VALUES ($1, $2, $3, $4, $5)
            RETURNING id, group_name, song_name, release_date, text, link, created_at, updated_at
        `
		err = tx.QueryRowContext(ctx, query, song.GroupName, song.SongName, song.ReleaseDate, song.Text, song.Link).Scan(
			&addedSong.ID, &addedSong.GroupName, &addedSong.SongName, &addedSong.ReleaseDate, &addedSong.Text, &addedSong.Link, &addedSong.CreatedAt, &addedSong.UpdatedAt,
		)
	} else {
		query = `
            INSERT INTO songs (group_name, song_name, release_date, text, link)
            VALUES ($1, $2, $3, $4, $5)
            RETURNING id, group_name, song_name, release_date, text, link, created_at, updated_at
        `
		err = s.conn.QueryRow(ctx, query, song.GroupName, song.SongName, song.ReleaseDate, song.Text, song.Link).Scan(
			&addedSong.ID, &addedSong.GroupName, &addedSong.SongName, &addedSong.ReleaseDate, &addedSong.Text, &addedSong.Link, &addedSong.CreatedAt, &addedSong.UpdatedAt,
		)
	}

	if err != nil {
		utils.Logger.Error("PgStorage.Create - queryRow failed", zap.Error(err))
		return nil, fmt.Errorf("PgStorage.Create - queryRow failed: %w", err)
	}
	return &addedSong, nil
}

func (s *PgStorage) GetByID(ctx context.Context, id int) (*models.Song, error) {
	query := `SELECT id, group_name, song_name, release_date, text, link, created_at, updated_at FROM songs WHERE id = $1`
	var song models.Song
	err := s.conn.QueryRow(ctx, query, id).Scan(
		&song.ID, &song.GroupName, &song.SongName, &song.ReleaseDate, &song.Text, &song.Link, &song.CreatedAt, &song.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrSongNotFound
		}
		utils.Logger.Error("PgStorage.GetByID - queryRow failed", zap.Error(err), zap.Int("id", id))
		return nil, fmt.Errorf("PgStorage.GetByID - queryRow failed: %w", err)
	}
	return &song, nil
}

func (s *PgStorage) List(ctx context.Context, filter *models.SongFilter, pagination *models.Pagination) ([]models.Song, error) {
	query := `SELECT id, group_name, song_name, release_date, text, link, created_at, updated_at FROM songs WHERE 1=1`
	var params []interface{}
	paramCount := 0

	if filter != nil {
		if filter.GroupName != nil && *filter.GroupName != "" {
			paramCount++
			query += fmt.Sprintf(" AND group_name ILIKE $%d", paramCount)
			params = append(params, "%"+*filter.GroupName+"%")
		}
		if filter.SongName != nil && *filter.SongName != "" {
			paramCount++
			query += fmt.Sprintf(" AND song_name ILIKE $%d", paramCount)
			params = append(params, "%"+*filter.SongName+"%")
		}
	}

	query += fmt.Sprintf(" LIMIT %d OFFSET %d", pagination.GetLimit(), pagination.GetOffset())

	rows, err := s.conn.Query(ctx, query, params...)
	if err != nil {
		utils.Logger.Error("PgStorage.List - query failed", zap.Error(err), zap.Any("filter", filter), zap.Any("pagination", pagination))
		return nil, fmt.Errorf("PgStorage.List - query failed: %w", err)
	}
	defer rows.Close()

	var songs []models.Song
	for rows.Next() {
		var song models.Song
		err := rows.Scan(
			&song.ID, &song.GroupName, &song.SongName, &song.ReleaseDate, &song.Text, &song.Link, &song.CreatedAt, &song.UpdatedAt,
		)
		if err != nil {
			utils.Logger.Error("PgStorage.List - rows.Scan failed", zap.Error(err))
			return nil, fmt.Errorf("PgStorage.List - rows.Scan failed: %w", err)
		}
		songs = append(songs, song)
	}

	if err := rows.Err(); err != nil {
		utils.Logger.Error("PgStorage.List - rows.Err failed", zap.Error(err))
		return nil, fmt.Errorf("PgStorage.List - rows.Err failed: %w", err)
	}

	return songs, nil
}

func (s *PgStorage) Update(ctx context.Context, song *models.Song) (*models.Song, error) {
	query := `
        UPDATE songs
        SET group_name = $1, song_name = $2, release_date = $3, text = $4, link = $5, updated_at = CURRENT_TIMESTAMP
        WHERE id = $6
        RETURNING id, group_name, song_name, release_date, text, link, created_at, updated_at
    `
	var updatedSong models.Song
	err := s.conn.QueryRow(
		ctx,
		query,
		&song.GroupName, &song.SongName, song.ReleaseDate, song.Text, song.Link, song.ID,
	).Scan(
		&updatedSong.ID, &updatedSong.GroupName, &updatedSong.SongName, &updatedSong.ReleaseDate, &updatedSong.Text, &updatedSong.Link, &updatedSong.CreatedAt, &updatedSong.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrSongNotFound
		}
		utils.Logger.Error("PgStorage.Update - queryRow failed", zap.Error(err), zap.Int("id", song.ID))
		return nil, fmt.Errorf("PgStorage.Update - queryRow failed: %w", err)
	}
	return &updatedSong, nil
}

func (s *PgStorage) Delete(ctx context.Context, id int) error {
	result, err := s.conn.Exec(ctx, "DELETE FROM songs WHERE id = $1", id)
	if err != nil {
		utils.Logger.Error("PgStorage.Delete - exec failed", zap.Error(err), zap.Int("id", id))
		return fmt.Errorf("PgStorage.Delete - exec failed: %w", err)
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return storage.ErrSongNotFound
	}
	return nil
}
