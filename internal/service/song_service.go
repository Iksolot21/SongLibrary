// internal/service/song_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"songlibrary/internal/lib/logger/utils"
	"songlibrary/internal/models"
	"songlibrary/internal/musicapi"
	"songlibrary/internal/storage"

	"go.uber.org/zap"
)

type SongService struct {
	storage        storage.SongStorage
	musicAPIClient *musicapi.MusicAPIClient
}

func NewSongService(storage storage.SongStorage, musicAPIClient *musicapi.MusicAPIClient) *SongService {
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
		return nil, fmt.Errorf("SongService.AddSong - GetSongDetailsFromAPI failed: %w", err)
	}

	var releaseDate *string
	if songDetails.ReleaseDate != "" {
		releaseDate = &songDetails.ReleaseDate
	}
	var text *string
	if songDetails.Text != "" {
		text = &songDetails.Text
	}
	var link *string
	if songDetails.Link != "" {
		link = &songDetails.Link
	}

	newSong := &models.Song{
		GroupName:   req.GroupName,
		SongName:    req.SongName,
		ReleaseDate: releaseDate,
		Text:        text,
		Link:        link,
	}

	addedSong, err := s.storage.Create(ctx, newSong)
	if err != nil {
		utils.Logger.Error("SongService.AddSong - storage.Create failed", zap.Error(err))
		return nil, fmt.Errorf("SongService.AddSong - storage.Create failed: %w", err)
	}

	utils.Logger.Info("SongService.AddSong - song added", zap.Int("song_id", addedSong.ID), zap.String("group", addedSong.GroupName), zap.String("song", addedSong.SongName))
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

func (s *SongService) GetSongText(ctx context.Context, id int) (*models.Song, error) {
	utils.Logger.Debug("SongService.GetSongText", zap.Int("id", id))

	song, err := s.storage.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrSongNotFound) {
			return nil, storage.ErrSongNotFound
		}
		utils.Logger.Error("SongService.GetSongText - storage.GetByID failed", zap.Error(err), zap.Int("id", id))
		return nil, fmt.Errorf("SongService.GetSongText - storage.GetByID failed: %w", err)
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
	utils.Logger.Info("SongService.UpdateSong - song updated", zap.Int("song_id", updatedSong.ID), zap.String("group", updatedSong.GroupName), zap.String("song", updatedSong.SongName))
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
