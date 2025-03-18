// internal/api/handlers/songs/songs_handlers_test.go
package songs_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"songlibrary/internal/api/handlers/songs"
	"songlibrary/internal/models"
	mock_service "songlibrary/internal/service/swagger" // Путь к сгенерированным mock файлам
	"songlibrary/internal/storage"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestAddSongHandler_Unit(t *testing.T) {
	testCases := []struct {
		name           string
		requestBody    string
		mockServiceFn  func(s *mock_service.MockSongService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Valid request",
			requestBody: `{"group": "Test Group", "song": "Test Song"}`,
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().AddSong(gomock.Any(), gomock.Any()).Return(&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song"}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":1,"group":"Test Group","song":"Test Song","releaseDate":null,"text":null,"link":null,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}\n`, // Adjust expected JSON
		},
		{
			name:           "Invalid request body",
			requestBody:    `invalid json`,
			mockServiceFn:  func(s *mock_service.MockSongService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request body"}\n`, // Adjust expected error JSON
		},
		{
			name:           "Missing group name",
			requestBody:    `{"song": "Test Song"}`,
			mockServiceFn:  func(s *mock_service.MockSongService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Group and song names are required"}\n`, // Adjust expected error JSON
		},
		{
			name:        "Service error",
			requestBody: `{"group": "Test Group", "song": "Test Song"}`,
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().AddSong(gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to add song"}\n`, // Adjust expected error JSON
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockSongService(ctrl)
			tc.mockServiceFn(mockService)

			handler := songs.NewSongHandlers(mockService)

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
		mockServiceFn  func(s *mock_service.MockSongService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "No filters, no pagination",
			queryParams: "",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().GetSongs(gomock.Any(), nil, gomock.Any()).Return([]models.Song{{ID: 1, GroupName: "Test Group", SongName: "Test Song"}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"group":"Test Group","song":"Test Song","releaseDate":null,"text":null,"link":null,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]\n`,
		},
		{
			name:        "Filter by group",
			queryParams: "?group=Test%20Group",
			mockServiceFn: func(s *mock_service.MockSongService) {
				filter := &models.SongFilter{GroupName: stringPointer("Test Group")}
				s.EXPECT().GetSongs(gomock.Any(), filter, gomock.Any()).Return([]models.Song{{ID: 1, GroupName: "Test Group", SongName: "Test Song"}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"group":"Test Group","song":"Test Song","releaseDate":null,"text":null,"link":null,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]\n`,
		},
		{
			name:        "Service error",
			queryParams: "",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().GetSongs(gomock.Any(), nil, gomock.Any()).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to get songs"}\n`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockSongService(ctrl)
			tc.mockServiceFn(mockService)

			handler := songs.NewSongHandlers(mockService)

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
		mockServiceFn  func(s *mock_service.MockSongService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Valid request",
			songID: "1",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().GetSongText(gomock.Any(), 1).Return(&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song", Text: stringPointer("Test Text")}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"group":"Test Group","song":"Test Song","releaseDate":null,"text":"Test Text","link":null,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}\n`,
		},
		{
			name:           "Invalid song ID",
			songID:         "invalid",
			mockServiceFn:  func(s *mock_service.MockSongService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid song ID"}\n`,
		},
		{
			name:   "Song not found",
			songID: "1",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().GetSongText(gomock.Any(), 1).Return(nil, storage.ErrSongNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Song not found"}\n`,
		},
		{
			name:   "Service error",
			songID: "1",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().GetSongText(gomock.Any(), 1).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to get song text"}\n`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockSongService(ctrl)
			tc.mockServiceFn(mockService)

			handler := songs.NewSongHandlers(mockService)
			req := httptest.NewRequest("GET", "/songs/"+tc.songID+"/text", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.songID}) // Set URL vars
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
		mockServiceFn  func(s *mock_service.MockSongService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Valid request",
			songID:      "1",
			requestBody: `{"id": 1, "group": "Updated Group", "song": "Updated Song"}`,
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().UpdateSong(gomock.Any(), gomock.Any()).Return(&models.Song{ID: 1, GroupName: "Updated Group", SongName: "Updated Song"}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"group":"Updated Group","song":"Updated Song","releaseDate":null,"text":null,"link":null,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}\n`,
		},
		{
			name:           "Invalid song ID",
			songID:         "invalid",
			requestBody:    `{"group": "Updated Group", "song": "Updated Song"}`,
			mockServiceFn:  func(s *mock_service.MockSongService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid song ID"}\n`,
		},
		{
			name:           "Invalid request body",
			songID:         "1",
			requestBody:    `invalid json`,
			mockServiceFn:  func(s *mock_service.MockSongService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request body"}\n`,
		},
		{
			name:        "Song not found",
			songID:      "1",
			requestBody: `{"id": 1, "group": "Updated Group", "song": "Updated Song"}`,
			mockServiceFn: func(s *mock_service.MockSongService) {
				updatedSong := models.Song{ID: 1, GroupName: "Updated Group", SongName: "Updated Song"} // Create a Song struct
				s.EXPECT().UpdateSong(gomock.Any(), gomock.Any()).Return(nil, storage.ErrSongNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Song not found"}\n`,
		},
		{
			name:        "Service error",
			songID:      "1",
			requestBody: `{"id": 1, "group": "Updated Group", "song": "Updated Song"}`,
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().UpdateSong(gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to update song"}\n`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockSongService(ctrl)
			tc.mockServiceFn(mockService)

			handler := songs.NewSongHandlers(mockService)
			req := httptest.NewRequest("PUT", "/songs/"+tc.songID, bytes.NewBufferString(tc.requestBody))
			req = mux.SetURLVars(req, map[string]string{"id": tc.songID}) // Set URL vars
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
		mockServiceFn  func(s *mock_service.MockSongService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Valid request",
			songID: "1",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().DeleteSong(gomock.Any(), 1).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   ``, // No body for 204
		},
		{
			name:           "Invalid song ID",
			songID:         "invalid",
			mockServiceFn:  func(s *mock_service.MockSongService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid song ID"}\n`,
		},
		{
			name:   "Song not found",
			songID: "1",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().DeleteSong(gomock.Any(), 1).Return(storage.ErrSongNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Song not found"}\n`,
		},
		{
			name:   "Service error",
			songID: "1",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().DeleteSong(gomock.Any(), 1).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to delete song"}\n`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockSongService(ctrl)
			tc.mockServiceFn(mockService)

			handler := songs.NewSongHandlers(mockService)
			req := httptest.NewRequest("DELETE", "/songs/"+tc.songID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.songID}) // Set URL vars
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

	mockService := mock_service.NewMockSongService(ctrl) // Mock service is not used in HealthCheckHandler

	handler := songs.NewSongHandlers(mockService)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheckHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func stringPointer(s string) *string {
	return &s
}
