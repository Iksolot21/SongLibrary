package songs_test

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"songlibrary/internal/api/handlers/songs"
	"songlibrary/internal/models"
	"songlibrary/internal/storage"
	mock_storage "songlibrary/internal/storage/mocks"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestAddSongHandler_Unit(t *testing.T) {
	testCases := []struct {
		name           string
		requestBody    string
		mockServiceFn  func(s *mock_storage.MockSongStorage)
		mockMusicAPIFn func(s *mock_storage.MockSongStorage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Valid request",
			requestBody: `{"group": "Test Group", "song": "Test Song"}`,
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				//s.EXPECT().AddSong(gomock.Any(), gomock.Any()).Return(&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song"}, nil)
				s.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song"}, nil)
			},
			mockMusicAPIFn: func(s *mock_storage.MockSongStorage) {},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":1,"group":"Test Group","song":"Test Song","releaseDate":null,"text":null,"link":null,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}\n`, // Adjust expected JSON
		},
		{
			name:           "Invalid request body",
			requestBody:    `invalid json`,
			mockServiceFn:  func(s *mock_storage.MockSongStorage) {},
			mockMusicAPIFn: func(s *mock_storage.MockSongStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request body"}\n`,
		},
		{
			name:           "Missing group name",
			requestBody:    `{"song": "Test Song"}`,
			mockServiceFn:  func(s *mock_storage.MockSongStorage) {},
			mockMusicAPIFn: func(s *mock_storage.MockSongStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Group and song names are required"}\n`,
		},
		{
			name:        "Service error",
			requestBody: `{"group": "Test Group", "song": "Test Song"}`,
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				//s.EXPECT().AddSong(gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))
				s.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))

			},
			mockMusicAPIFn: func(s *mock_storage.MockSongStorage) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to add song"}\n`,
		},
		//{
		//	name:        "Service external API error",
		//	requestBody: `{"group": "Test Group", "song": "Test Song"}`,
		//	mockServiceFn: func(s *mock_service.MockSongService) {
		//		s.EXPECT().AddSong(gomock.Any(), gomock.Any()).Return(nil, service.ErrExternalAPI)
		//	},
		//	expectedStatus: http.StatusServiceUnavailable,
		//	expectedBody:   `{"error":"Failed to add song"}\n`, // Adjust expected error JSON
		//},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mock_storage.NewMockSongStorage(ctrl)
			tc.mockServiceFn(mockStorage)

			handler := songs.NewSongHandlers(nil)

			req := httptest.NewRequest("POST", "/songs", bytes.NewBufferString(tc.requestBody))
			w := httptest.NewRecorder()

			handler.AddSongHandler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.JSONEq(t, tc.expectedBody, w.Body.String())
		})
	}
}

func TestGetSongsHandler_Unit(t *testing.T) {
	testCases := []struct {
		name           string
		queryParams    string
		mockServiceFn  func(s *mock_storage.MockSongStorage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "No filters, no pagination",
			queryParams: "",
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().List(gomock.Any(), nil, gomock.Any()).Return([]models.Song{{ID: 1, GroupName: "Test Group", SongName: "Test Song"}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"group":"Test Group","song":"Test Song","releaseDate":null,"text":null,"link":null,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]\n`,
		},
		{
			name:        "Filter by group",
			queryParams: "?group=Test%20Group",
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				filter := &models.SongFilter{GroupName: stringPointer("Test Group")}
				s.EXPECT().List(gomock.Any(), filter, gomock.Any()).Return([]models.Song{{ID: 1, GroupName: "Test Group", SongName: "Test Song"}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"group":"Test Group","song":"Test Song","releaseDate":null,"text":null,"link":null,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]\n`,
		},
		{
			name:        "Service error",
			queryParams: "",
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().List(gomock.Any(), nil, gomock.Any()).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to get songs"}\n`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mock_storage.NewMockSongStorage(ctrl)
			tc.mockServiceFn(mockStorage)

			handler := songs.NewSongHandlers(nil)

			req := httptest.NewRequest("GET", "/songs"+tc.queryParams, nil)
			w := httptest.NewRecorder()

			handler.GetSongsHandler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.JSONEq(t, tc.expectedBody, w.Body.String())
		})
	}
}

func TestGetSongTextHandler_Unit(t *testing.T) {
	testCases := []struct {
		name           string
		songID         string
		queryParams    string
		mockServiceFn  func(s *mock_storage.MockSongStorage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Valid request",
			songID:      "1",
			queryParams: "",
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().GetByID(gomock.Any(), 1).Return(&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song", Text: sqlStringPointer("Test Text")}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"group":"Test Group","song":"Test Song","releaseDate":null,"text":"Test Text","link":null,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}\n`,
		},
		{
			name:        "Valid request with pagination",
			songID:      "1",
			queryParams: "?page=1&pageSize=1",
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().GetByID(gomock.Any(), 1).Return(&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song", Text: sqlStringPointer("Verse1\n\nVerse2\n\nVerse3")}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"group":"Test Group","song":"Test Song","releaseDate":null,"text":"Verse1","link":null,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}\n`,
		},
		{
			name:        "Request with pagination no content",
			songID:      "1",
			queryParams: "?page=10&pageSize=1",
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().GetByID(gomock.Any(), 1).Return(&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song", Text: sqlStringPointer("")}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"group":"Test Group","song":"Test Song","releaseDate":null,"text":"","link":null,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}\n`,
		},
		{
			name:           "Invalid song ID",
			songID:         "invalid",
			queryParams:    "",
			mockServiceFn:  func(s *mock_storage.MockSongStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid song ID"}\n`,
		},
		{
			name:        "Song not found",
			songID:      "1",
			queryParams: "",
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().GetByID(gomock.Any(), 1).Return(nil, storage.ErrSongNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Song not found"}\n`,
		},
		{
			name:        "Service error",
			songID:      "1",
			queryParams: "",
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().GetByID(gomock.Any(), 1).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to get song text"}\n`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mock_storage.NewMockSongStorage(ctrl)
			tc.mockServiceFn(mockStorage)

			handler := songs.NewSongHandlers(nil)
			req := httptest.NewRequest("GET", "/songs/"+tc.songID+"/text"+tc.queryParams, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.songID})
			w := httptest.NewRecorder()

			handler.GetSongTextHandler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.JSONEq(t, tc.expectedBody, w.Body.String())
		})
	}
}

func TestUpdateSongHandler_Unit(t *testing.T) {
	testCases := []struct {
		name           string
		songID         string
		requestBody    string
		mockServiceFn  func(s *mock_storage.MockSongStorage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Valid request",
			songID:      "1",
			requestBody: `{"id": 1, "group": "Updated Group", "song": "Updated Song"}`,
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().Update(gomock.Any(), gomock.Any()).Return(&models.Song{ID: 1, GroupName: "Updated Group", SongName: "Updated Song"}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"group":"Updated Group","song":"Updated Song","releaseDate":null,"text":null,"link":null,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}\n`,
		},
		{
			name:           "Invalid song ID",
			songID:         "invalid",
			requestBody:    `{"group": "Updated Group", "song": "Updated Song"}`,
			mockServiceFn:  func(s *mock_storage.MockSongStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid song ID"}\n`,
		},
		{
			name:           "Invalid request body",
			songID:         "1",
			requestBody:    `invalid json`,
			mockServiceFn:  func(s *mock_storage.MockSongStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request body"}\n`,
		},
		{
			name:        "Song not found",
			songID:      "1",
			requestBody: `{"id": 1, "group": "Updated Group", "song": "Updated Song"}`,
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				//updatedSong := models.Song{ID: 1, GroupName: "Updated Group", SongName: "Updated Song"} // Create a Song struct
				s.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, storage.ErrSongNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Song not found"}\n`,
		},
		{
			name:        "Service error",
			songID:      "1",
			requestBody: `{"id": 1, "group": "Updated Group", "song": "Updated Song"}`,
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				//updatedSong := models.Song{ID: 1, GroupName: "Updated Group", SongName: "Updated Song"} // Create a Song struct
				s.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to update song"}\n`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mock_storage.NewMockSongStorage(ctrl)
			tc.mockServiceFn(mockStorage)

			handler := songs.NewSongHandlers(nil)
			req := httptest.NewRequest("PUT", "/songs/"+tc.songID, bytes.NewBufferString(tc.requestBody))
			req = mux.SetURLVars(req, map[string]string{"id": tc.songID})
			w := httptest.NewRecorder()

			handler.UpdateSongHandler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.JSONEq(t, tc.expectedBody, w.Body.String())
		})
	}
}

func TestDeleteSongHandler_Unit(t *testing.T) {
	testCases := []struct {
		name           string
		songID         string
		mockServiceFn  func(s *mock_storage.MockSongStorage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Valid request",
			songID: "1",
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().Delete(gomock.Any(), 1).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   ``,
		},
		{
			name:           "Invalid song ID",
			songID:         "invalid",
			mockServiceFn:  func(s *mock_storage.MockSongStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid song ID"}\n`,
		},
		{
			name:   "Song not found",
			songID: "1",
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().Delete(gomock.Any(), 1).Return(storage.ErrSongNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Song not found"}\n`,
		},
		{
			name:   "Service error",
			songID: "1",
			mockServiceFn: func(s *mock_storage.MockSongStorage) {
				s.EXPECT().Delete(gomock.Any(), 1).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to delete song"}\n`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mock_storage.NewMockSongStorage(ctrl)
			tc.mockServiceFn(mockStorage)

			handler := songs.NewSongHandlers(nil)
			req := httptest.NewRequest("DELETE", "/songs/"+tc.songID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.songID})
			w := httptest.NewRecorder()

			handler.DeleteSongHandler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.JSONEq(t, tc.expectedBody, w.Body.String())
		})
	}
}

func TestHealthCheckHandler_Unit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//mockService := mock_service.NewMockSongService(ctrl) // Mock service is not used in HealthCheckHandler

	handler := songs.NewSongHandlers(nil)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheckHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
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
