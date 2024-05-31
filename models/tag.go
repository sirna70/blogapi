package models

type Tag struct {
	ID    int    `json:"id"`
	Label string `json:"label" validate:"required"`
	Posts int64  `json:"posts_id"`
}
