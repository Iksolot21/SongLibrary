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
	"time"

	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/mock_service.go -package=mocks songlibrary/internal/service SongService

var (
	ErrExternalAPI = errors.New("external API error")
)

type SongService interface {
	AddSong(ctx context.Context, req *models.AddSongRequest) (*models.Song, error)
	GetSongs(ctx context.Context, filter *models.SongFilter, pagination *models.Pagination) ([]models.Song, error)
	GetSongText(ctx context.Context, id int, pagination *models.Pagination) (*models.Song, error)
	UpdateSong(ctx context.Context, song *models.Song) (*models.Song, error)
	DeleteSong(ctx context.Context, id int) error
}

type songService struct {
	storage        storage.SongStorage
	musicAPIClient musicapi.MusicAPI
}

func NewSongService(storage storage.SongStorage, musicAPIClient musicapi.MusicAPI) SongService {
	return &songService{
		storage:        storage,
		musicAPIClient: musicAPIClient,
	}
}

func (s *songService) AddSong(ctx context.Context, req *models.AddSongRequest) (*models.Song, error) {
	utils.Logger.Debug("SongService.AddSong", zap.String("group", req.GroupName), zap.String("song", req.SongName))

	songDetails, err := s.musicAPIClient.GetSongDetailsFromAPI(req.GroupName, req.SongName)
	if err != nil {
		utils.Logger.Error("SongService.AddSong - GetSongDetailsFromAPI failed", zap.Error(err))
		return nil, fmt.Errorf("SongService.AddSong - GetSongDetailsFromAPI failed: %w", ErrExternalAPI)
	}

	const maxTextLength = 65535
	if len(songDetails.Text) > maxTextLength {
		return nil, fmt.Errorf("text length exceeds maximum allowed length (%d)", maxTextLength)
	}

	const maxReleaseDateLength = 255
	if len(songDetails.ReleaseDate) > maxReleaseDateLength {
		return nil, fmt.Errorf("releaseDate length exceeds maximum allowed length (%d)", maxReleaseDateLength)
	}

	var nullReleaseDate sql.NullString
	if songDetails.ReleaseDate != "" {
		parsedTime, parseErr := time.Parse("02.01.2006", songDetails.ReleaseDate)
		if parseErr != nil {
			utils.Logger.Warn("SongService.AddSong - Failed to parse releaseDate from API, using empty date", zap.Error(parseErr), zap.String("releaseDate", songDetails.ReleaseDate))
			nullReleaseDate = sql.NullString{String: "", Valid: false}
		} else {
			nullReleaseDate = sql.NullString{String: parsedTime.Format("2006-01-02"), Valid: true}
		}
	} else {
		nullReleaseDate = sql.NullString{String: "", Valid: false}
	}

	nullText := sql.NullString{String: songDetails.Text, Valid: songDetails.Text != ""}
	nullLink := sql.NullString{String: songDetails.Link, Valid: songDetails.Link != ""}

	newSong := &models.Song{
		GroupName:   req.GroupName,
		SongName:    req.SongName,
		ReleaseDate: nullReleaseDate,
		Text:        nullText,
		Link:        nullLink,
	}

	addedSong, err := s.storage.Create(ctx, newSong, nil)
	if err != nil {
		utils.Logger.Error("SongService.AddSong - storage.Create failed", zap.Error(err))
		return nil, fmt.Errorf("SongService.AddSong - storage.Create failed: %w", err)
	}

	utils.Logger.Info("SongService.AddSong - song added", zap.Int("song_id", addedSong.ID), zap.String("group", req.GroupName), zap.String("song", req.SongName))
	return addedSong, nil
}

func (s *songService) GetSongs(ctx context.Context, filter *models.SongFilter, pagination *models.Pagination) ([]models.Song, error) {
	utils.Logger.Debug("SongService.GetSongs", zap.Any("filter", filter), zap.Any("pagination", pagination))

	songs, err := s.storage.List(ctx, filter, pagination)
	if err != nil {
		utils.Logger.Error("SongService.GetSongs - storage.List failed", zap.Error(err), zap.Any("filter", filter), zap.Any("pagination", pagination))
		return nil, fmt.Errorf("SongService.GetSongs - storage.List failed: %w", err)
	}
	return songs, nil
}

func (s *songService) GetSongText(ctx context.Context, id int, pagination *models.Pagination) (*models.Song, error) {
	utils.Logger.Debug("SongService.GetSongText", zap.Int("id", id), zap.Any("pagination", pagination))

	song, err := s.storage.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrSongNotFound) {
			return nil, storage.ErrSongNotFound
		}
		utils.Logger.Error("SongService.GetSongText - storage.GetByID failed", zap.Error(err), zap.Int("id", id))
		return nil, fmt.Errorf("SongService.GetSongText - storage.GetByID failed: %w", err)
	}

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

func (s *songService) UpdateSong(ctx context.Context, song *models.Song) (*models.Song, error) {
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

func (s *songService) DeleteSong(ctx context.Context, id int) error {
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
