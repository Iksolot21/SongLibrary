package models

import (
	"database/sql"
	"time"

	_ "github.com/swaggo/swag"
)

type Song struct {
	ID        int    `json:"id"`
	GroupName string `json:"group"`
	SongName  string `json:"song"`
	// swagger:strfmt date-time
	ReleaseDate sql.NullString `json:"releaseDate" swaggertype:"string"`
	// swagger:strfmt string
	Text sql.NullString `json:"text" swaggertype:"string"`
	// swagger:strfmt uri
	Link      sql.NullString `json:"link" swaggertype:"string"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
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
