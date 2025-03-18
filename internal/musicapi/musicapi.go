// internal/musicapi/musicapi.go
package musicapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"songlibrary/internal/lib/logger/utils"
	"songlibrary/internal/models"

	"go.uber.org/zap"
)

type MusicAPIClient struct {
	baseURL string
	client  *http.Client
}

func NewMusicAPIClient(baseURL string) *MusicAPIClient {
	return &MusicAPIClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (api *MusicAPIClient) GetSongDetailsFromAPI(group, song string) (*models.SongDetailFromAPI, error) {
	apiURL := api.baseURL
	if apiURL == "" {
		return nil, fmt.Errorf("API_URL not configured in .env file")
	}

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
