// internal/models/song.go
package models

import "time"

type Song struct {
	ID          int       `json:"id"`
	GroupName   string    `json:"group"`
	SongName    string    `json:"song"`
	ReleaseDate *string   `json:"releaseDate"`
	Text        *string   `json:"text"`
	Link        *string   `json:"link"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type SongDetailFromAPI struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type AddSongRequest struct {
	GroupName string `json:"group"`
	SongName  string `json:"song"`
}

type SongFilter struct {
	GroupName *string
	SongName  *string
}
