// tests/songlibrary_test.go
package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"songlibrary/config"
	"songlibrary/internal/api/handlers/songs"
	"songlibrary/internal/lib/logger/utils"
	"songlibrary/internal/models"
	"songlibrary/internal/musicapi"
	"songlibrary/internal/service"
	"songlibrary/internal/storage"
	"songlibrary/internal/storage/postgres"
)

var (
	testDBConnStr  string
	testServer     *httptest.Server
	testRouter     *mux.Router
	pgStorage      storage.SongStorage
	musicAPIClient *musicapi.MusicAPIClient
	songHandlers   *songs.SongHandlers
	songService    *service.SongService
)

func setupTestEnvironment(t *testing.T) func() {
	godotenv.Load("../.env")

	cfg, err := config.LoadConfig()
	require.NoError(t, err, "Failed to load config")

	testDBConnStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName+"_test")

	conn, err := pgx.Connect(context.Background(), testDBConnStr)
	require.NoError(t, err, "Failed to connect to test database")

	pgStorage = postgres.NewPgStorage(conn)
	musicAPIClient = musicapi.NewMusicAPIClient(cfg.APIURL)
	songService = service.NewSongService(pgStorage, musicAPIClient)
	songHandlers = songs.NewSongHandlers(songService)

	testRouter = mux.NewRouter()
	testRouter.HandleFunc("/health", songHandlers.HealthCheckHandler).Methods("GET")
	testRouter.HandleFunc("/songs", songHandlers.GetSongsHandler).Methods("GET")
	testRouter.HandleFunc("/songs", songHandlers.AddSongHandler).Methods("POST")
	testRouter.HandleFunc("/songs/{id}/text", songHandlers.GetSongTextHandler).Methods("GET")
	testRouter.HandleFunc("/songs/{id}", songHandlers.UpdateSongHandler).Methods("PUT")
	testRouter.HandleFunc("/songs/{id}", songHandlers.DeleteSongHandler).Methods("DELETE")

	testServer = httptest.NewServer(testRouter)

	// Clean up database after tests
	return func() {
		conn.Close(context.Background())
		cleanupTestData(t)
		testServer.Close()
	}
}

func cleanupTestData(t *testing.T) {
	conn, err := pgx.Connect(context.Background(), testDBConnStr)
	require.NoError(t, err, "Failed to connect to test database for cleanup")
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "DELETE FROM songs")
	require.NoError(t, err, "Failed to cleanup test data")
}

func executeRequest(t *testing.T, method, path string, body string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, testServer.URL+path, bytes.NewBufferString(body))
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, req)
	return recorder
}

func TestHealthCheckHandler_Integration(t *testing.T) {
	teardown := setupTestEnvironment(t)
	defer teardown()

	recorder := executeRequest(t, "GET", "/health", "")
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "OK", recorder.Body.String())
}

func TestGetSongsHandler_Integration(t *testing.T) {
	teardown := setupTestEnvironment(t)
	defer teardown()

	// Add test data directly to DB
	addTestData(t)

	recorder := executeRequest(t, "GET", "/songs", "")
	assert.Equal(t, http.StatusOK, recorder.Code)

	var songs []models.Song
	err := json.Unmarshal(recorder.Body.Bytes(), &songs)
	require.NoError(t, err, "Failed to unmarshal response body")
	assert.NotEmpty(t, songs, "Expected songs in response")
}

func TestAddSongHandler_Integration(t *testing.T) {
	teardown := setupTestEnvironment(t)
	defer teardown()

	requestBody := `{"group": "Integration Test Group", "song": "Integration Test Song"}`
	recorder := executeRequest(t, "POST", "/songs", requestBody)
	assert.Equal(t, http.StatusCreated, recorder.Code)

	var song models.Song
	err := json.Unmarshal(recorder.Body.Bytes(), &song)
	require.NoError(t, err, "Failed to unmarshal response body")
	assert.Equal(t, "Integration Test Group", song.GroupName)
	assert.Equal(t, "Integration Test Song", song.SongName)

	// Verify song exists in DB
	fetchedSong, err := pgStorage.GetByID(context.Background(), song.ID)
	require.NoError(t, err, "Failed to fetch song from DB")
	assert.Equal(t, song.GroupName, fetchedSong.GroupName)
	assert.Equal(t, song.SongName, fetchedSong.SongName)
}

func TestGetSongTextHandler_Integration(t *testing.T) {
	teardown := setupTestEnvironment(t)
	defer teardown()

	// Add test data with text
	testSong := addTestDataWithText(t)

	recorder := executeRequest(t, "GET", "/songs/"+strconv.Itoa(testSong.ID)+"/text", "")
	assert.Equal(t, http.StatusOK, recorder.Code)

	var song models.Song
	err := json.Unmarshal(recorder.Body.Bytes(), &song)
	require.NoError(t, err, "Failed to unmarshal response body")
	assert.Equal(t, testSong.Text, song.Text)
}

func TestUpdateSongHandler_Integration(t *testing.T) {
	teardown := setupTestEnvironment(t)
	defer teardown()

	// Add test data to update
	testSong := addTestData(t)[0]
	updatedGroupName := "Updated Group Name"
	updatedSongName := "Updated Song Name"
	updateRequestBody := fmt.Sprintf(`{"id": %d, "group": "%s", "song": "%s"}`, testSong.ID, updatedGroupName, updatedSongName)

	recorder := executeRequest(t, "PUT", "/songs/"+strconv.Itoa(testSong.ID), updateRequestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var updatedSong models.Song
	err := json.Unmarshal(recorder.Body.Bytes(), &updatedSong)
	require.NoError(t, err, "Failed to unmarshal response body")
	assert.Equal(t, updatedGroupName, updatedSong.GroupName)
	assert.Equal(t, updatedSongName, updatedSong.SongName)

	// Verify song is updated in DB
	fetchedSong, err := pgStorage.GetByID(context.Background(), testSong.ID)
	require.NoError(t, err, "Failed to fetch song from DB")
	assert.Equal(t, updatedGroupName, fetchedSong.GroupName)
	assert.Equal(t, updatedSongName, fetchedSong.SongName)
}

func TestDeleteSongHandler_Integration(t *testing.T) {
	teardown := setupTestEnvironment(t)
	defer teardown()

	// Add test data to delete
	testSong := addTestData(t)[0]

	recorder := executeRequest(t, "DELETE", "/songs/"+strconv.Itoa(testSong.ID), "")
	assert.Equal(t, http.StatusNoContent, recorder.Code)

	// Verify song is deleted from DB
	_, err := pgStorage.GetByID(context.Background(), testSong.ID)
	assert.ErrorIs(t, err, storage.ErrSongNotFound, "Expected song to be deleted")
}

func addTestData(t *testing.T) []models.Song {
	songsToAdd := []models.Song{
		{GroupName: "Test Group 1", SongName: "Test Song 1"},
		{GroupName: "Test Group 2", SongName: "Test Song 2"},
	}
	addedSongs := make([]models.Song, len(songsToAdd))

	for i, song := range songsToAdd {
		addedSong, err := pgStorage.Create(context.Background(), &song, nil)
		require.NoError(t, err, "Failed to add test song to DB")
		addedSongs[i] = *addedSong
	}
	return addedSongs
}

func addTestDataWithText(t *testing.T) models.Song {
	songToAdd := models.Song{
		GroupName: "Text Group",
		SongName:  "Text Song",
		Text:      sql.NullString{String: "Test Song Text", Valid: true},
	}
	addedSong, err := pgStorage.Create(context.Background(), &songToAdd, nil)
	require.NoError(t, err, "Failed to add test song with text to DB")
	return *addedSong
}

func TestMain(m *testing.M) {
	if err := utils.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer utils.Logger.Sync()
	utils.Logger.Info("Starting Integration Tests", zap.String("test_suite", "integration"))
	exitCode := m.Run()
	utils.Logger.Info("Integration Tests Finished", zap.Int("exit_code", exitCode))
	os.Exit(exitCode)
}
