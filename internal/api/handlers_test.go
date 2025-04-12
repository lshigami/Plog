package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lshigami/Plog/internal/config"
	mock_sqlc "github.com/lshigami/Plog/internal/db/mock"
	"github.com/lshigami/Plog/internal/db/sqlc"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Định nghĩa struct response tương ứng với handler
type PostResponse struct {
	ID             int32     `json:"id"`
	UserID         int32     `json:"user_id"`
	AuthorUsername string    `json:"author_username"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Định nghĩa struct lỗi
type HTTPError struct {
	Error string `json:"error"`
}

func setupTestServer(t *testing.T, store sqlc.Querier) *Server {
	fakeConfig := config.Config{
		DatabaseURL:         "postgres",
		ServerPort:          "8080",
		AccessTokenDuration: time.Minute,
		JWTSecret:           "a_very_secret_key_should_be_longer_and_random",
	}
	server := NewServer(fakeConfig, store)
	return server
}

func setupGinTest() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/posts", nil)
	return c, w
}

func TestListPostsAPI(t *testing.T) {
	mockPosts := []sqlc.ListPostsRow{
		{
			ID:             1,
			UserID:         10,
			AuthorUsername: "testuser",
			Title:          "First Post",
			Content:        "Content 1",
			CreatedAt:      pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
			UpdatedAt:      pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
		},
		{
			ID:             2,
			UserID:         10,
			AuthorUsername: "testuser",
			Title:          "Second Post",
			Content:        "Content 2",
			CreatedAt:      pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
			UpdatedAt:      pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
		},
	}

	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mock_sqlc.NewMockQuerier(ctrl)
		server := setupTestServer(t, mockStore)
		c, recorder := setupGinTest()

		limit := int32(10)
		offset := int32(0)
		params := sqlc.ListPostsParams{Limit: limit, Offset: offset}

		mockStore.EXPECT().
			ListPosts(gomock.Any(), params).
			Times(1).
			Return(mockPosts, nil)

		c.Request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/posts?limit=%d&offset=%d", limit, offset), nil)
		server.ListPosts(c)

		require.Equal(t, http.StatusOK, recorder.Code)

		var responseBody []PostResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		require.Len(t, responseBody, len(mockPosts))
		require.Equal(t, mockPosts[0].Title, responseBody[0].Title)
		require.Equal(t, mockPosts[1].ID, responseBody[1].ID)

		// Sửa lại phần so sánh thời gian
		require.WithinDuration(t, mockPosts[0].CreatedAt.Time, responseBody[0].CreatedAt, time.Second)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mock_sqlc.NewMockQuerier(ctrl)
		server := setupTestServer(t, mockStore)
		c, recorder := setupGinTest()

		limit := int32(5)
		offset := int32(0)
		params := sqlc.ListPostsParams{Limit: limit, Offset: offset}
		dbError := fmt.Errorf("some database error")

		mockStore.EXPECT().
			ListPosts(gomock.Any(), params).
			Times(1).
			Return(nil, dbError)

		c.Request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/posts?limit=%d&offset=%d", limit, offset), nil)
		server.ListPosts(c)

		require.Equal(t, http.StatusInternalServerError, recorder.Code)

		var errorResponse HTTPError
		err := json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
		require.NoError(t, err)
		require.Contains(t, errorResponse.Error, "Failed to list posts")
	})

	t.Run("InvalidQueryParams", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mock_sqlc.NewMockQuerier(ctrl)
		server := setupTestServer(t, mockStore)
		c, recorder := setupGinTest()

		c.Request, _ = http.NewRequest(http.MethodGet, "/posts?limit=-1&offset=0", nil)
		server.ListPosts(c)

		require.Equal(t, http.StatusBadRequest, recorder.Code)

		var errorResponse HTTPError
		err := json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
		require.NoError(t, err)
		require.Contains(t, errorResponse.Error, "Invalid input")
	})
}
