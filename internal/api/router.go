package api

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lshigami/Plog/docs" // Import the docs package to register Swagger docs
	"github.com/lshigami/Plog/internal/config"
	"github.com/lshigami/Plog/internal/db/sqlc"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(store sqlc.Querier, cfg config.Config) *gin.Engine {

	router := gin.Default() //  logger & recovery middleware

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	server := NewServer(cfg, store)
	server.router = router

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.POST("/register", server.RegisterUser)
	router.POST("/login", server.LoginUser)

	// Post routes (Public GET)
	postRoutes := router.Group("/posts")
	{
		postRoutes.GET("", server.ListPosts)
		postRoutes.GET("/:id", server.GetPost)
	}

	// Authenticated routes
	authRoutes := router.Group("/")
	authRoutes.Use(AuthMiddleware(server.tokenMaker))
	{

		authRoutes.POST("/posts", server.CreatePost)
		authRoutes.PUT("/posts/:id", server.UpdatePost)

	}

	return router
}
