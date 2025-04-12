-- internal/db/query.sql

-- name: CreateUser :one
INSERT INTO users (username, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: CreatePost :one
INSERT INTO posts (user_id, title, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPostByID :one
SELECT p.*, u.username as author_username
FROM posts p
JOIN users u ON p.user_id = u.id
WHERE p.id = $1 LIMIT 1;

-- name: ListPosts :many
SELECT p.*, u.username as author_username
FROM posts p
JOIN users u ON p.user_id = u.id
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2; -- For pagination

-- name: UpdatePost :one
UPDATE posts
SET title = $2, content = $3, updated_at = NOW()
WHERE id = $1 AND user_id = $4 -- Ensure user owns the post
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts
WHERE id = $1 AND user_id = $2; -- Ensure user owns the post