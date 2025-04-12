package api

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lshigami/Plog/internal/auth"
	"github.com/lshigami/Plog/internal/db/sqlc"
)

type RegisterUserRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserResponse struct {
	ID        int32     `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

func newUserResponse(user sqlc.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Time,
	}
}

type LoginUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginUserResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type CreatePostRequest struct {
	Title   string `json:"title" binding:"required,min=3,max=255"`
	Content string `json:"content" binding:"required"`
}

type ListPostsRequest struct {
	Limit  int32 `form:"limit,default=10" binding:"min=1,max=100"`
	Offset int32 `form:"offset,default=0" binding:"min=0"`
}

type UpdatePostRequest struct {
	Title   string `json:"title" binding:"required,min=3,max=255"`
	Content string `json:"content" binding:"required"`
}

// RegisterUser godoc
// @Summary Register a new user
// @Description Register a new user with username and password
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body RegisterUserRequest true "User registration details"
// @Success 201 {object} UserResponse "User created successfully"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 409 {object} map[string]string "Username already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /register [post]
func (server *Server) RegisterUser(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	arg := sqlc.CreateUserParams{
		Username:     req.Username,
		PasswordHash: hashedPassword,
	}

	user, err := server.store.CreateUser(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	rsp := newUserResponse(user)
	c.JSON(http.StatusCreated, rsp)
}

// LoginUser godoc
// @Summary Login a user
// @Description Authenticate a user and return an access token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body LoginUserRequest true "User login credentials"
// @Success 200 {object} LoginUserResponse "Login successful"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Invalid username or password"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /login [post]
func (server *Server) LoginUser(c *gin.Context) {
	var req LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	user, err := server.store.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user: " + err.Error()})
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(user.ID, user.Username, server.config.AccessTokenDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access token"})
		return
	}

	rsp := LoginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user),
	}
	c.JSON(http.StatusOK, rsp)

}

// CreatePost godoc
// @Summary Create a new post
// @Description Create a new blog post
// @Tags posts
// @Accept json
// @Produce json
// @Param request body CreatePostRequest true "Post details"
// @Success 201 {object} SwaggerPost "Post created successfully"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /posts [post]
func (server *Server) CreatePost(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	userID, ok := c.Get(UserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	arg := sqlc.CreatePostParams{
		UserID:  userID.(int32),
		Title:   req.Title,
		Content: req.Content,
	}

	post, err := server.store.CreatePost(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post: " + err.Error()})
		return
	}

	fullPost, err := server.store.GetPostByID(c.Request.Context(), post.ID)
	if err != nil {

		log.Printf("Warning: could not fetch full post details after creation: %v", err)
		c.JSON(http.StatusCreated, post)
		return
	}

	c.JSON(http.StatusCreated, fullPost)
}

// GetPost godoc
// @Summary Get a post by ID
// @Description Get details of a specific post by its ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} SwaggerPost "Post details"
// @Failure 400 {object} map[string]string "Invalid post ID format"
// @Failure 404 {object} map[string]string "Post not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /posts/{id} [get]
func (server *Server) GetPost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32) // Convert string to int32
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID format"})
		return
	}

	post, err := server.store.GetPostByID(c.Request.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get post: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

// ListPosts godoc
// @Summary List posts
// @Description Get a list of posts with pagination
// @Tags posts
// @Accept json
// @Produce json
// @Param limit query int true "Limit" minimum(1) maximum(100)
// @Param offset query int true "Offset" minimum(0)
// @Success 200 {array} SwaggerPost "List of posts"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /posts [get]
func (server *Server) ListPosts(c *gin.Context) {
	var req ListPostsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	arg := sqlc.ListPostsParams{
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	posts, err := server.store.ListPosts(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list posts: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, posts)
}

// UpdatePost godoc
// @Summary Update a post
// @Description Update a post's title and content
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param request body UpdatePostRequest true "Updated post details"
// @Success 200 {object} SwaggerPost "Updated post"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Post not found or no permission"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /posts/{id} [put]
func (server *Server) UpdatePost(c *gin.Context) {
	idStr := c.Param("id")
	postID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID format"})
		return
	}
	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	userID := c.MustGet(UserIDKey).(int)

	arg := sqlc.UpdatePostParams{
		ID:      int32(postID),
		Title:   req.Title,
		Content: req.Content,
		UserID:  int32(userID),
	}

	post, err := server.store.UpdatePost(c.Request.Context(), arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Could be post not found OR user doesn't own it
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found or you don't have permission to update it"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post: " + err.Error()})
		return
	}

	fullPost, err := server.store.GetPostByID(c.Request.Context(), post.ID)
	if err != nil {
		log.Printf("Warning: could not fetch full post details after update: %v", err)
		c.JSON(http.StatusOK, post)
		return
	}

	c.JSON(http.StatusOK, fullPost)
}
