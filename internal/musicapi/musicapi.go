package musicapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"songlibrary/internal/lib/logger/utils"
	"songlibrary/internal/models"
	"strings"

	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/mock_musicapi.go -package=mocks songlibrary/internal/musicapi MusicAPI

type MusicAPI interface {
	GetSongDetailsFromAPI(group string, song string) (*models.SongDetailFromAPI, error)
}

type MusicAPIClient struct {
	baseURL     string
	client      *http.Client
	useMockData bool
}

func NewMusicAPIClient(baseURL string) *MusicAPIClient {
	useMockData := baseURL == ""
	if useMockData {
		utils.Logger.Info("MusicAPIClient initialized in mock data mode because API_URL is not configured.")
	} else {
		utils.Logger.Info("MusicAPIClient initialized with API_URL", zap.String("url", baseURL))
	}

	return &MusicAPIClient{
		baseURL:     baseURL,
		client:      &http.Client{},
		useMockData: useMockData,
	}
}

func (api *MusicAPIClient) GetSongDetailsFromAPI(group string, song string) (*models.SongDetailFromAPI, error) {
	if api.useMockData {
		utils.Logger.Debug("MusicAPIClient is in mock data mode. Returning mock data.")
		mockSongDetails := new(models.SongDetailFromAPI)
		mockSongDetails.ReleaseDate = "2023-10-27"
		mockSongDetails.Text = fmt.Sprintf("Mock Text: This is a sample verse for '%s' by '%s'.\n\n(Data from Mock MusicAPIClient)", song, group)
		mockSongDetails.Link = "https://www.youtube.com/watch?v=mockExample"
		return mockSongDetails, nil
	}

	apiURL := api.baseURL
	if apiURL == "" {
		return nil, fmt.Errorf("API_URL not configured and mock data mode is not active, which should not happen")
	}

	apiURL = strings.TrimSuffix(apiURL, "/") + "/info"

	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API_URL: %w", err)
	}

	query := u.Query()
	query.Set("group", group)
	query.Set("song", song)
	u.RawQuery = query.Encode()

	utils.Logger.Debug("Calling external API", zap.String("url", u.String()))

	resp, err := api.client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to call external API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external API returned error: %s", resp.Status)
	}

	var songDetails models.SongDetailFromAPI
	if err := json.NewDecoder(resp.Body).Decode(&songDetails); err != nil {
		return nil, fmt.Errorf("failed to decode external API response: %w", err)
	}

	utils.Logger.Debug("External API response", zap.Any("details", songDetails))
	return &songDetails, nil
}
