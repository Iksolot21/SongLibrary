package songs_test

import (
	"bytes"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"songlibrary/internal/api/handlers/songs"
	"songlibrary/internal/lib/logger/utils"
	"songlibrary/internal/models"
	mock_service "songlibrary/internal/service/mocks"
	"songlibrary/internal/storage"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	if err := utils.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	exitCode := m.Run()
	utils.Logger.Sync()
	os.Exit(exitCode)
}

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
				s.EXPECT().AddSong(gomock.Any(), gomock.Any()).Return(
					&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song"},
					nil,
				)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":1,"group":"Test Group","song":"Test Song","releaseDate":{"String":"","Valid":false},"text":{"String":"","Valid":false},"link":{"String":"","Valid":false},"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}`, // Исправлено
		},
		{
			name:           "Invalid request body",
			requestBody:    `invalid json`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request body"}`,
		},
		{
			name:           "Missing group name",
			requestBody:    `{"song": "Test Song"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Group and song names are required"}`,
		},
		{
			name:        "Service error",
			requestBody: `{"group": "Test Group", "song": "Test Song"}`,
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().AddSong(gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to add song"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockSongService(ctrl)
			if tc.mockServiceFn != nil {
				tc.mockServiceFn(mockService)
			}

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
				s.EXPECT().GetSongs(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					[]models.Song{{ID: 1, GroupName: "Test Group", SongName: "Test Song"}},
					nil,
				)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"group":"Test Group","song":"Test Song","releaseDate":{"String":"","Valid":false},"text":{"String":"","Valid":false},"link":{"String":"","Valid":false},"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]`, // Исправлено
		},
		{
			name:        "Filter by group",
			queryParams: "?group=Test%20Group",
			mockServiceFn: func(s *mock_service.MockSongService) {
				filter := &models.SongFilter{GroupName: stringPointer("Test Group")}
				s.EXPECT().GetSongs(gomock.Any(), gomock.Eq(filter), gomock.Any()).Return(
					[]models.Song{{ID: 1, GroupName: "Test Group", SongName: "Test Song"}},
					nil,
				)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"group":"Test Group","song":"Test Song","releaseDate":{"String":"","Valid":false},"text":{"String":"","Valid":false},"link":{"String":"","Valid":false},"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]`, // Исправлено
		},
		{
			name:        "Service error",
			queryParams: "",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().GetSongs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to get songs"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockSongService(ctrl)
			if tc.mockServiceFn != nil {
				tc.mockServiceFn(mockService)
			}

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
		queryParams    string
		mockServiceFn  func(s *mock_service.MockSongService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Valid request",
			songID:      "1",
			queryParams: "",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().GetSongText(gomock.Any(), gomock.Eq(1), gomock.Any()).Return(
					&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song", Text: sqlStringPointer("Test Text")},
					nil,
				)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"group":"Test Group","song":"Test Song","releaseDate":{"String":"","Valid":false},"text":{"String":"Test Text","Valid":true},"link":{"String":"","Valid":false},"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}`, // Исправлено
		},
		{
			name:        "Valid request with pagination",
			songID:      "1",
			queryParams: "?page=1&pageSize=1",
			mockServiceFn: func(s *mock_service.MockSongService) {
				pagination := models.NewPagination(1, 1)
				s.EXPECT().GetSongText(gomock.Any(), gomock.Eq(1), gomock.Eq(pagination)).Return(
					&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song", Text: sqlStringPointer("Verse1")},
					nil,
				)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"group":"Test Group","song":"Test Song","releaseDate":{"String":"","Valid":false},"text":{"String":"Verse1","Valid":true},"link":{"String":"","Valid":false},"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}`, // Исправлено
		},
		{
			name:        "Request with pagination no content",
			songID:      "1",
			queryParams: "?page=10&pageSize=1",
			mockServiceFn: func(s *mock_service.MockSongService) {
				pagination := models.NewPagination(10, 1)
				s.EXPECT().GetSongText(gomock.Any(), gomock.Eq(1), gomock.Eq(pagination)).Return(
					&models.Song{ID: 1, GroupName: "Test Group", SongName: "Test Song", Text: sqlStringPointer("")},
					nil,
				)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"group":"Test Group","song":"Test Song","releaseDate":{"String":"","Valid":false},"text":{"String":"","Valid":true},"link":{"String":"","Valid":false},"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}`, // Исправлено, text valid true because sql.NullString is present, even if string is empty
		},
		{
			name:           "Invalid song ID",
			songID:         "invalid",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid song ID"}`,
		},
		{
			name:        "Song not found",
			songID:      "1",
			queryParams: "",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().GetSongText(gomock.Any(), gomock.Eq(1), gomock.Any()).Return(nil, storage.ErrSongNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Song not found"}`,
		},
		{
			name:        "Service error",
			songID:      "1",
			queryParams: "",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().GetSongText(gomock.Any(), gomock.Eq(1), gomock.Any()).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to get song text"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockSongService(ctrl)
			if tc.mockServiceFn != nil {
				tc.mockServiceFn(mockService)
			}

			handler := songs.NewSongHandlers(mockService)
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
		mockServiceFn  func(s *mock_service.MockSongService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Valid request",
			songID:      "1",
			requestBody: `{"id": 1, "group": "Updated Group", "song": "Updated Song"}`,
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().UpdateSong(gomock.Any(), gomock.Any()).Return(
					&models.Song{ID: 1, GroupName: "Updated Group", SongName: "Updated Song"},
					nil,
				)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"group":"Updated Group","song":"Updated Song","releaseDate":{"String":"","Valid":false},"text":{"String":"","Valid":false},"link":{"String":"","Valid":false},"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}`, // Исправлено
		},
		{
			name:           "Invalid song ID",
			songID:         "invalid",
			requestBody:    `{"group": "Updated Group", "song": "Updated Song"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid song ID"}`,
		},
		{
			name:           "Invalid request body",
			songID:         "1",
			requestBody:    `invalid json`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request body"}`,
		},
		{
			name:        "Song not found",
			songID:      "1",
			requestBody: `{"id": 1, "group": "Updated Group", "song": "Updated Song"}`,
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().UpdateSong(gomock.Any(), gomock.Any()).Return(nil, storage.ErrSongNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Song not found"}`,
		},
		{
			name:        "Service error",
			songID:      "1",
			requestBody: `{"id": 1, "group": "Updated Group", "song": "Updated Song"}`,
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().UpdateSong(gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to update song"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockSongService(ctrl)
			if tc.mockServiceFn != nil {
				tc.mockServiceFn(mockService)
			}

			handler := songs.NewSongHandlers(mockService)
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
		mockServiceFn  func(s *mock_service.MockSongService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Valid request",
			songID: "1",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().DeleteSong(gomock.Any(), gomock.Eq(1)).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   ``,
		},
		{
			name:           "Invalid song ID",
			songID:         "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid song ID"}`,
		},
		{
			name:   "Song not found",
			songID: "1",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().DeleteSong(gomock.Any(), gomock.Eq(1)).Return(storage.ErrSongNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Song not found"}`,
		},
		{
			name:   "Service error",
			songID: "1",
			mockServiceFn: func(s *mock_service.MockSongService) {
				s.EXPECT().DeleteSong(gomock.Any(), gomock.Eq(1)).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to delete song"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockSongService(ctrl)
			if tc.mockServiceFn != nil {
				tc.mockServiceFn(mockService)
			}

			handler := songs.NewSongHandlers(mockService)
			req := httptest.NewRequest("DELETE", "/songs/"+tc.songID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.songID})
			w := httptest.NewRecorder()

			handler.DeleteSongHandler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.name == "Valid request" {
				assert.Empty(t, w.Body.String())
			} else {
				assert.JSONEq(t, tc.expectedBody, w.Body.String())
			}
		})
	}
}

func TestHealthCheckHandler_Unit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
