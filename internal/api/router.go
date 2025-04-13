package api

import (
	"net/http"
	"os"
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
	// staticAssetsPath = "/app/static/assets" // Ví dụ
	// urlAssetsPrefix  = "/assets"          // Ví dụ
)

func SetupRouter(store sqlc.Querier, cfg config.Config) *gin.Engine {

	router := gin.Default()

	// --- CORS Configuration ---
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000"} // Điều chỉnh nếu cần
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

	// 1. Phục vụ các thư mục assets (CSS, JS, etc.)
	//    Đảm bảo đường dẫn nguồn (http.Dir) khớp với vị trí trong container
	router.StaticFS("/static", http.Dir(filepath.Join(staticRootPath, "static"))) // Ví dụ: /app/static/static
	router.StaticFS("/assets", http.Dir(filepath.Join(staticRootPath, "assets"))) // Ví dụ: /app/static/assets (Nếu có)
	// Thêm các thư mục static khác nếu cần

	// 2. *** THAY ĐỔI QUAN TRỌNG: Xử lý route gốc ("/") một cách tường minh ***
	//    Thay vì dùng StaticFileFS, dùng GET handler
	// router.StaticFileFS("/", filepath.Join(staticRootPath, staticIndexFile), http.Dir(staticRootPath)) // <<< XÓA DÒNG NÀY
	router.GET("/", func(c *gin.Context) {
		// Tạo một filesystem ảo gốc tại staticRootPath
		fs := http.Dir(staticRootPath)
		// Phục vụ file index.html từ gốc của filesystem đó
		c.FileFromFS(staticIndexFile, fs)
	})

	// --- SPA Catch-all Route ---
	// Logic này giữ nguyên, nó sẽ không được gọi cho "/" nữa vì đã có handler GET ở trên
	router.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet &&
			!strings.HasPrefix(c.Request.URL.Path, "/api/") &&
			!strings.HasPrefix(c.Request.URL.Path, "/static/") &&
			!strings.HasPrefix(c.Request.URL.Path, "/assets/") && // Thêm kiểm tra này nếu dùng /assets
			!strings.HasPrefix(c.Request.URL.Path, "/swagger/") {

			// Kiểm tra file có tồn tại vật lý không trước khi trả về index.html
			filePath := filepath.Join(staticRootPath, filepath.Clean(c.Request.URL.Path))
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				c.FileFromFS(staticIndexFile, http.Dir(staticRootPath)) // Trả về index.html
				return
			}
			// Nếu file tồn tại nhưng không được serve bởi StaticFS, để Gin xử lý 404
		}
		// Trả về 404 cho các trường hợp khác
		// c.JSON(http.StatusNotFound, HTTPError{Error: "Resource not found"}) // Hoặc 404 mặc định của Gin
	})

	return router
}

// ... (Phần còn lại của code: các struct, hàm handler, middleware...)
