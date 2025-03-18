package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"songlibrary/internal/lib/logger/utils"
	"songlibrary/internal/models"
	"songlibrary/internal/musicapi"
	"songlibrary/internal/storage"
	"strings"

	"go.uber.org/zap"
)

var (
	ErrExternalAPI = errors.New("external API error")
)

type SongService struct {
	storage        storage.SongStorage
	musicAPIClient musicapi.MusicAPI
}

func NewSongService(storage storage.SongStorage, musicAPIClient musicapi.MusicAPI) *SongService {
	return &SongService{
		storage:        storage,
		musicAPIClient: musicAPIClient,
	}
}

func (s *SongService) AddSong(ctx context.Context, req *models.AddSongRequest) (*models.Song, error) {
	utils.Logger.Debug("SongService.AddSong", zap.String("group", req.GroupName), zap.String("song", req.SongName))

	songDetails, err := s.musicAPIClient.GetSongDetailsFromAPI(req.GroupName, req.SongName)
	if err != nil {
		utils.Logger.Error("SongService.AddSong - GetSongDetailsFromAPI failed", zap.Error(err))
		return nil, fmt.Errorf("SongService.AddSong - GetSongDetailsFromAPI failed: %w", ErrExternalAPI)
	}

	// Validate data lengths
	const maxTextLength = 65535
	if len(songDetails.Text) > maxTextLength {
		return nil, fmt.Errorf("text length exceeds maximum allowed length (%d)", maxTextLength)
	}

	const maxReleaseDateLength = 255
	if len(songDetails.ReleaseDate) > maxReleaseDateLength {
		return nil, fmt.Errorf("releaseDate length exceeds maximum allowed length (%d)", maxReleaseDateLength)
	}

	// Start a transaction
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				utils.Logger.Error("Transaction rollback failed", zap.Error(rollbackErr))
			}
		}
	}()

	nullReleaseDate := sql.NullString{String: songDetails.ReleaseDate, Valid: songDetails.ReleaseDate != ""}
	nullText := sql.NullString{String: songDetails.Text, Valid: songDetails.Text != ""}
	nullLink := sql.NullString{String: songDetails.Link, Valid: songDetails.Link != ""}

	newSong := &models.Song{
		GroupName:   req.GroupName,
		SongName:    req.SongName,
		ReleaseDate: nullReleaseDate,
		Text:        nullText,
		Link:        nullLink,
	}

	addedSong, err := s.storage.Create(ctx, newSong, tx)
	if err != nil {
		utils.Logger.Error("SongService.AddSong - storage.Create failed", zap.Error(err))
		return nil, fmt.Errorf("SongService.AddSong - storage.Create failed: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	utils.Logger.Info("SongService.AddSong - song added", zap.Int("song_id", addedSong.ID), zap.String("group", req.GroupName), zap.String("song", req.SongName))
	return addedSong, nil
}

func (s *SongService) GetSongs(ctx context.Context, filter *models.SongFilter, pagination *models.Pagination) ([]models.Song, error) {
	utils.Logger.Debug("SongService.GetSongs", zap.Any("filter", filter), zap.Any("pagination", pagination))

	songs, err := s.storage.List(ctx, filter, pagination)
	if err != nil {
		utils.Logger.Error("SongService.GetSongs - storage.List failed", zap.Error(err), zap.Any("filter", filter), zap.Any("pagination", pagination))
		return nil, fmt.Errorf("SongService.GetSongs - storage.List failed: %w", err)
	}
	return songs, nil
}

func (s *SongService) GetSongText(ctx context.Context, id int, pagination *models.Pagination) (*models.Song, error) {
	utils.Logger.Debug("SongService.GetSongText", zap.Int("id", id), zap.Any("pagination", pagination))

	song, err := s.storage.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrSongNotFound) {
			return nil, storage.ErrSongNotFound
		}
		utils.Logger.Error("SongService.GetSongText - storage.GetByID failed", zap.Error(err), zap.Int("id", id))
		return nil, fmt.Errorf("SongService.GetSongText - storage.GetByID failed: %w", err)
	}

	// Implement verse pagination
	if song.Text.Valid {
		verses := strings.Split(song.Text.String, "\n\n")
		startIndex := pagination.GetOffset()
		endIndex := startIndex + pagination.GetLimit()

		if startIndex > len(verses) {
			song.Text = sql.NullString{String: "", Valid: false}
		} else {
			if endIndex > len(verses) {
				endIndex = len(verses)
			}
			paginatedVerses := verses[startIndex:endIndex]
			song.Text = sql.NullString{String: strings.Join(paginatedVerses, "\n\n"), Valid: true}
		}
	}

	return song, nil
}

func (s *SongService) UpdateSong(ctx context.Context, song *models.Song) (*models.Song, error) {
	utils.Logger.Debug("SongService.UpdateSong", zap.Int("id", song.ID), zap.String("group", song.GroupName), zap.String("song", song.SongName))

	updatedSong, err := s.storage.Update(ctx, song)
	if err != nil {
		if errors.Is(err, storage.ErrSongNotFound) {
			return nil, storage.ErrSongNotFound
		}
		utils.Logger.Error("SongService.UpdateSong - storage.Update failed", zap.Error(err), zap.Int("id", song.ID))
		return nil, fmt.Errorf("SongService.UpdateSong - storage.Update failed: %w", err)
	}
	utils.Logger.Info("SongService.UpdateSong - song updated", zap.Int("song_id", updatedSong.ID), zap.String("group", song.GroupName), zap.String("song", song.SongName))
	return updatedSong, nil
}

func (s *SongService) DeleteSong(ctx context.Context, id int) error {
	utils.Logger.Debug("SongService.DeleteSong", zap.Int("id", id))

	err := s.storage.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrSongNotFound) {
			return storage.ErrSongNotFound
		}
		utils.Logger.Error("SongService.DeleteSong - storage.Delete failed", zap.Error(err), zap.Int("id", id))
		return fmt.Errorf("SongService.DeleteSong - storage.Delete failed: %w", err)
	}
	utils.Logger.Info("SongService.DeleteSong - song deleted", zap.Int("song_id", id))
	return nil
}
