package service_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"songlibrary/internal/models"
	mock_musicapi "songlibrary/internal/musicapi/mocks"
	"songlibrary/internal/service"
	"songlibrary/internal/storage"
	mock_storage "songlibrary/internal/storage/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSongService_AddSong(t *testing.T) {
	testCases := []struct {
		name           string
		request        *models.AddSongRequest
		mockStorageFn  func(s *mock_storage.MockSongStorage)
		mockMusicAPIFn func(m *mock_musicapi.MockMusicAPI)
		expectError    bool
	}{
		{
			name: "Valid request",
			request: &models.AddSongRequest{
				GroupName: "Test Group",
				SongName:  "Test Song",
			},
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().BeginTx(gomock.Any()).Return(&sql.Tx{}, nil)
				s.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song"}, nil)
			},
			mockMusicAPIFn: func(m *mock_musicapi.MockMusicAPI) {
				m.EXPECT().GetSongDetailsFromAPI("Test Group", "Test Song").Return(&models.SongDetailFromAPI{Text: "Test Text", ReleaseDate: "2023-01-01", Link: "http://test.link"}, nil)
			},
			expectError: false,
		},
		{
			name: "MusicAPIClient error",
			request: &models.AddSongRequest{
				GroupName: "Test Group",
				SongName:  "Test Song",
			},
			mockStorageFn: func(s *mock_storage.MockSongStorage) {},
			mockMusicAPIFn: func(m *mock_musicapi.MockMusicAPI) {
				m.EXPECT().GetSongDetailsFromAPI("Test Group", "Test Song").Return(nil, errors.New("music api error"))
			},
			expectError: true,
		},
		{
			name: "Storage error",
			request: &models.AddSongRequest{
				GroupName: "Test Group",
				SongName:  "Test Song",
			},
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().BeginTx(gomock.Any()).Return(&sql.Tx{}, nil)
				s.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("storage error"))
			},
			mockMusicAPIFn: func(m *mock_musicapi.MockMusicAPI) {
				m.EXPECT().GetSongDetailsFromAPI("Test Group", "Test Song").Return(&models.SongDetailFromAPI{Text: "Test Text", ReleaseDate: "2023-01-01", Link: "http://test.link"}, nil)
			},
			expectError: true,
		},
		{
			name: "Text length exceeds limit",
			request: &models.AddSongRequest{
				GroupName: "Test Group",
				SongName:  "Test Song",
			},
			mockStorageFn: func(s *mock_storage.MockSongStorage) {},
			mockMusicAPIFn: func(m *mock_musicapi.MockMusicAPI) {
				longText := strings.Repeat("A", 65536)
				m.EXPECT().GetSongDetailsFromAPI("Test Group", "Test Song").Return(&models.SongDetailFromAPI{Text: longText, ReleaseDate: "2023-01-01", Link: "http://test.link"}, nil)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mock_storage.NewMockSongStorage(ctrl)
			mockMusicAPIClient := mock_musicapi.NewMockMusicAPI(ctrl)

			tc.mockStorageFn(mockStorage)
			tc.mockMusicAPIFn(mockMusicAPIClient)

			serviceInstance := service.NewSongService(mockStorage, mockMusicAPIClient)

			_, err := serviceInstance.AddSong(context.Background(), tc.request)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSongService_GetSongs(t *testing.T) {
	testCases := []struct {
		name          string
		filter        *models.SongFilter
		pagination    *models.Pagination
		mockStorageFn func(s *mock_storage.MockSongStorage)
		expectError   bool
	}{
		{
			name:       "No filter, no pagination",
			filter:     nil,
			pagination: models.NewPagination(1, 10),
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().List(gomock.Any(), nil, gomock.Any()).Return([]models.Song{{ID: 1, GroupName: "Test Group", SongName: "Test Song"}}, nil)
			},
			expectError: false,
		},
		{
			name: "Filter by group",
			filter: &models.SongFilter{
				GroupName: stringPointer("Test Group"),
			},
			pagination: models.NewPagination(1, 10),
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				filter := &models.SongFilter{GroupName: stringPointer("Test Group")}
				s.EXPECT().List(gomock.Any(), filter, gomock.Any()).Return([]models.Song{{ID: 1, GroupName: "Test Group", SongName: "Test Song"}}, nil)
			},
			expectError: false,
		},
		{
			name:       "Storage error",
			filter:     nil,
			pagination: models.NewPagination(1, 10),
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().List(gomock.Any(), nil, gomock.Any()).Return(nil, errors.New("storage error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mock_storage.NewMockSongStorage(ctrl)
			tc.mockStorageFn(mockStorage)

			serviceInstance := service.NewSongService(mockStorage, nil)

			_, err := serviceInstance.GetSongs(context.Background(), tc.filter, tc.pagination)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSongService_GetSongText(t *testing.T) {
	testCases := []struct {
		name          string
		songID        int
		pagination    *models.Pagination
		mockStorageFn func(s *mock_storage.MockSongStorage)
		expectError   bool
		expectedSong  *models.Song
	}{
		{
			name:       "Valid request",
			songID:     1,
			pagination: models.NewPagination(1, 10),
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().GetByID(gomock.Any(), 1).Return(&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song", Text: sql.NullString{String: "Test Text", Valid: true}}, nil)
			},
			expectError:  false,
			expectedSong: &models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song", Text: sql.NullString{String: "Test Text", Valid: true}},
		},
		{
			name:       "Song not found",
			songID:     1,
			pagination: models.NewPagination(1, 10),
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().GetByID(gomock.Any(), 1).Return(nil, storage.ErrSongNotFound)
			},
			expectError:  true,
			expectedSong: nil,
		},
		{
			name:       "Storage error",
			songID:     1,
			pagination: models.NewPagination(1, 10),
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().GetByID(gomock.Any(), 1).Return(nil, errors.New("storage error"))
			},
			expectError:  true,
			expectedSong: nil,
		},
		{
			name:       "Valid request with pagination",
			songID:     1,
			pagination: models.NewPagination(1, 1),
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().GetByID(gomock.Any(), 1).Return(&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song", Text: sqlStringPointer("Verse1\n\nVerse2\n\nVerse3")}, nil)
			},
			expectError:  false,
			expectedSong: &models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song", Text: sqlStringPointer("Verse1")},
		},
		{
			name:       "Request with pagination no content",
			songID:     1,
			pagination: models.NewPagination(10, 1),
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().GetByID(gomock.Any(), 1).Return(&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song", Text: sqlStringPointer("Verse1\n\nVerse2\n\nVerse3")}, nil)
			},
			expectError:  false,
			expectedSong: &models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song", Text: sql.NullString{String: "", Valid: false}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mock_storage.NewMockSongStorage(ctrl)
			tc.mockStorageFn(mockStorage)

			serviceInstance := service.NewSongService(mockStorage, nil)

			song, err := serviceInstance.GetSongText(context.Background(), tc.songID, tc.pagination)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedSong.Text.String, song.Text.String)
			}
		})
	}
}

func TestSongService_UpdateSong(t *testing.T) {
	testCases := []struct {
		name          string
		songToUpdate  *models.Song
		mockStorageFn func(s *mock_storage.MockSongStorage)
		expectError   bool
	}{
		{
			name: "Valid request",
			songToUpdate: &models.Song{
				ID:        1,
				GroupName: "Updated Group",
				SongName:  "Updated Song",
			},
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().Update(gomock.Any(), gomock.Any()).Return(&models.Song{ID: 1, GroupName: "Updated Group", SongName: "Updated Song"}, nil)
			},
			expectError: false,
		},
		{
			name: "Song not found",
			songToUpdate: &models.Song{
				ID:        1,
				GroupName: "Updated Group",
				SongName:  "Updated Song",
			},
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, storage.ErrSongNotFound)
			},
			expectError: true,
		},
		{
			name: "Storage error",
			songToUpdate: &models.Song{
				ID:        1,
				GroupName: "Updated Group",
				SongName:  "Updated Song",
			},
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("storage error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mock_storage.NewMockSongStorage(ctrl)
			tc.mockStorageFn(mockStorage)

			serviceInstance := service.NewSongService(mockStorage, nil)

			_, err := serviceInstance.UpdateSong(context.Background(), tc.songToUpdate)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSongService_DeleteSong(t *testing.T) {
	testCases := []struct {
		name          string
		songID        int
		mockStorageFn func(s *mock_storage.MockSongStorage)
		expectError   bool
	}{
		{
			name:   "Valid request",
			songID: 1,
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().Delete(gomock.Any(), 1).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "Song not found",
			songID: 1,
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().Delete(gomock.Any(), 1).Return(storage.ErrSongNotFound)
			},
			expectError: true,
		},
		{
			name:   "Storage error",
			songID: 1,
			mockStorageFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().Delete(gomock.Any(), 1).Return(errors.New("storage error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mock_storage.NewMockSongStorage(ctrl)
			tc.mockStorageFn(mockStorage)

			serviceInstance := service.NewSongService(mockStorage, nil)

			err := serviceInstance.DeleteSong(context.Background(), tc.songID)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func stringPointer(s string) *string {
	return &s
}

func sqlStringPointer(s string) sql.NullString {
	sqlString := sql.NullString{
		String: s,
		Valid:  true,
	}
	return sqlString
}
