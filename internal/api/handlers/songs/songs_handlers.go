package songs

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"songlibrary/internal/lib/logger/utils"
	"songlibrary/internal/lib/response"
	"songlibrary/internal/models"
	"songlibrary/internal/service"
	"songlibrary/internal/storage"
)

type SongHandlers struct {
	songService service.SongService
}

func NewSongHandlers(songService service.SongService) *SongHandlers {
	return &SongHandlers{
		songService: songService,
	}
}

// @Summary Get songs with filtering and pagination
// @Description Get songs with optional filters for group and song name, and pagination.
// @Tags songs
// @Produce json
// @Param group query string false "Filter by group name"
// @Param song query string false "Filter by song name"
// @Param page query int false "Page number for pagination" default(1)
// @Param pageSize query int false "Number of songs per page" default(10)
// @Success 200 {array} models.Song
// @Router /songs [get]
// @swaggo:operation GET /songs getSongs
func (h *SongHandlers) GetSongsHandler(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("GetSongsHandler called")

	queryParams := r.URL.Query()
	page, _ := strconv.Atoi(queryParams.Get("page"))
	pageSize, _ := strconv.Atoi(queryParams.Get("pageSize"))

	pagination := models.NewPagination(page, pageSize)

	filter := &models.SongFilter{
		GroupName: stringPointer(queryParams.Get("group")),
		SongName:  stringPointer(queryParams.Get("song")),
	}

	songs, err := h.songService.GetSongs(r.Context(), filter, pagination)
	if err != nil {
		utils.Logger.Error("GetSongsHandler - songService.GetSongs failed", zap.Error(err), zap.Any("filter", filter), zap.Any("pagination", pagination))
		response.Error(w, http.StatusInternalServerError, "Failed to get songs")
		return
	}

	response.JSON(w, http.StatusOK, songs)
	utils.Logger.Debug("GetSongsHandler - songs retrieved", zap.Int("count", len(songs)))
}

// @Summary Add a new song
// @Description Add a new song to the library, fetching details from external API.
// @Tags songs
// @Accept json
// @Produce json
// @Param body body models.AddSongRequest true "Song details to add"
// @Success 201 {object} models.Song
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /songs [post]
// @swaggo:operation POST /songs addSong
func (h *SongHandlers) AddSongHandler(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("AddSongHandler called")
	var req models.AddSongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Logger.Warn("AddSongHandler - invalid request body", zap.Error(err))
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.GroupName == "" || req.SongName == "" {
		utils.Logger.Warn("AddSongHandler - group and song names are required")
		response.Error(w, http.StatusBadRequest, "Group and song names are required")
		return
	}

	addedSong, err := h.songService.AddSong(r.Context(), &req)
	if err != nil {
		utils.Logger.Error("AddSongHandler - songService.AddSong failed", zap.Error(err))
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrExternalAPI) {
			statusCode = http.StatusServiceUnavailable
		}
		response.Error(w, statusCode, "Failed to add song")
		return
	}

	response.JSON(w, http.StatusCreated, addedSong)
	utils.Logger.Info("AddSongHandler - song added successfully", zap.Int("song_id", addedSong.ID), zap.String("group", addedSong.GroupName), zap.String("song", addedSong.SongName))
}

// @Summary Get song text by ID with pagination
// @Description Get the text of a song by its ID, with pagination for verses.
// @Tags songs
// @Produce json
// @Param id path int true "Song ID"
// @Param page query int false "Page number for verses" default(1)
// @Param pageSize query int false "Number of verses per page" default(1)
// @Success 200 {object} models.Song
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /songs/{id}/text [get]
// @swaggo:operation GET /songs/{id}/text getSongText
func (h *SongHandlers) GetSongTextHandler(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("GetSongTextHandler called")
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Logger.Warn("GetSongTextHandler - invalid song ID", zap.Error(err), zap.String("id", idStr))
		response.Error(w, http.StatusBadRequest, "Invalid song ID")
		return
	}

	queryParams := r.URL.Query()
	page, _ := strconv.Atoi(queryParams.Get("page"))
	pageSize, _ := strconv.Atoi(queryParams.Get("pageSize"))

	pagination := models.NewPagination(page, pageSize)

	song, err := h.songService.GetSongText(r.Context(), id, pagination)
	if err != nil {
		if errors.Is(err, storage.ErrSongNotFound) {
			response.Error(w, http.StatusNotFound, "Song not found")
			return
		}
		utils.Logger.Error("GetSongTextHandler - songService.GetSongText failed", zap.Error(err), zap.Int("id", id))
		response.Error(w, http.StatusInternalServerError, "Failed to get song text")
		return
	}

	response.JSON(w, http.StatusOK, song)
	utils.Logger.Debug("GetSongTextHandler - song text retrieved", zap.Int("song_id", song.ID))
}

// @Summary Update song by ID
// @Description Update an existing song's details.
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Param body body models.Song true "Song details to update"
// @Success 200 {object} models.Song
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /songs/{id} [put]
// @swaggo:operation PUT /songs/{id} updateSong
func (h *SongHandlers) UpdateSongHandler(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("UpdateSongHandler called")
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Logger.Warn("UpdateSongHandler - invalid song ID", zap.Error(err), zap.String("id", idStr))
		response.Error(w, http.StatusBadRequest, "Invalid song ID")
		return
	}

	var updatedSongData models.Song
	if err := json.NewDecoder(r.Body).Decode(&updatedSongData); err != nil {
		utils.Logger.Warn("UpdateSongHandler - invalid request body", zap.Error(err))
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	updatedSongData.ID = id

	updatedSong, err := h.songService.UpdateSong(r.Context(), &updatedSongData)
	if err != nil {
		if errors.Is(err, storage.ErrSongNotFound) {
			response.Error(w, http.StatusNotFound, "Song not found")
			return
		}
		utils.Logger.Error("UpdateSongHandler - songService.UpdateSong failed", zap.Error(err), zap.Int("id", id))
		response.Error(w, http.StatusInternalServerError, "Failed to update song")
		return
	}

	response.JSON(w, http.StatusOK, updatedSong)
	utils.Logger.Info("UpdateSongHandler - song updated successfully", zap.Int("song_id", updatedSong.ID), zap.String("group", updatedSong.GroupName), zap.String("song", updatedSong.SongName))
}

// @Summary Delete song by ID
// @Description Delete a song from the library.
// @Tags songs
// @Produce json
// @Param id path int true "Song ID"
// @Success 204 "No Content"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /songs/{id} [delete]
// @swaggo:operation DELETE /songs/{id} deleteSong
func (h *SongHandlers) DeleteSongHandler(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("DeleteSongHandler called")
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Logger.Warn("DeleteSongHandler - invalid song ID", zap.Error(err), zap.String("id", idStr))
		response.Error(w, http.StatusBadRequest, "Invalid song ID")
		return
	}

	err = h.songService.DeleteSong(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrSongNotFound) {
			response.Error(w, http.StatusNotFound, "Song not found")
			return
		}
		utils.Logger.Error("DeleteSongHandler - songService.DeleteSong failed", zap.Error(err), zap.Int("id", id))
		response.Error(w, http.StatusInternalServerError, "Failed to delete song")
		return
	}

	w.WriteHeader(http.StatusNoContent)
	utils.Logger.Info("DeleteSongHandler - song deleted successfully", zap.Int("song_id", id))
}

// HealthCheckHandler godoc
// @Summary Show the status of server.
// @Description Get the status of server.
// @Tags root
// @Produce plain
// @Success 200 {string} string "OK"
// @Router /health [get]
// @swaggo:operation GET /health healthCheck
func (h *SongHandlers) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func stringPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
