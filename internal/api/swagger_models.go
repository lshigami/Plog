package api

import "time"

// SwaggerPost represents a blog post for Swagger documentation
// This is a duplicate of sqlc.Post but with standard Go types
// @Description A blog post
type SwaggerPost struct {
	ID        int32     `json:"id"`
	UserID    int32     `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
}
