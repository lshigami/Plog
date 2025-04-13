package api

import (
	"net/http"
	"path/filepath"
	"strings"

	// Đảm bảo time được import nếu dùng trong CORS
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lshigami/Plog/docs"
	"github.com/lshigami/Plog/internal/config"
	"github.com/lshigami/Plog/internal/db/sqlc"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	staticRootPath  = "/app/static" // Thư mục static trong container
	staticIndexFile = "index.html"
)

func SetupRouter(store sqlc.Querier, cfg config.Config) *gin.Engine {
	// Use ReleaseMode for production
	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// --- CORS Configuration ---
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000", "*"} // Allow all origins for testing
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// --- Server Instance ---
	server := NewServer(cfg, store)
	server.router = router

	// --- API Routes (/api/v1) ---
	apiV1 := router.Group("/api/v1")
	{
		// Swagger
		apiV1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		// Health
		apiV1.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
		// Auth
		apiV1.POST("/register", server.RegisterUser)
		apiV1.POST("/login", server.LoginUser)
		// Posts (Public)
		postRoutes := apiV1.Group("/posts")
		{
			postRoutes.GET("", server.ListPosts)
			postRoutes.GET("/:id", server.GetPost)
		}
		// Posts (Authenticated)
		authRoutes := apiV1.Group("/")
		authRoutes.Use(AuthMiddleware(server.tokenMaker)) // Đảm bảo AuthMiddleware đúng
		{
			authRoutes.POST("/posts", server.CreatePost)
			authRoutes.PUT("/posts/:id", server.UpdatePost)
			// authRoutes.DELETE("/posts/:id", server.DeletePost) // Nếu có
		}
	}

	// --- Static Frontend Files Serving ---
	// Serve static files from the React build directory
	router.StaticFile("/", filepath.Join(staticRootPath, "index.html"))
	router.Static("/static", filepath.Join(staticRootPath, "static"))
	
	// Serve other static assets that might be at the root level
	router.StaticFile("/favicon.ico", filepath.Join(staticRootPath, "favicon.ico"))
	router.StaticFile("/manifest.json", filepath.Join(staticRootPath, "manifest.json"))
	router.StaticFile("/asset-manifest.json", filepath.Join(staticRootPath, "asset-manifest.json"))

	// --- SPA Catch-all Route ---
	// Handle all other routes for the SPA
	router.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet &&
			!strings.HasPrefix(c.Request.URL.Path, "/api/") &&
			!strings.HasPrefix(c.Request.URL.Path, "/static/") &&
			!strings.HasPrefix(c.Request.URL.Path, "/swagger/") {

			// Always serve index.html for client-side routing
			c.File(filepath.Join(staticRootPath, "index.html"))
			return
		}
		// Return 404 for API routes or other non-SPA routes
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
	})

	return router
}

// ... (Phần còn lại của code: các struct, hàm handler, middleware...)
