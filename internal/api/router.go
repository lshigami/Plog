package api

import (
	"net/http"
	"os"            // Thêm import os
	"path/filepath" // Thêm import path/filepath
	"strings"       // Thêm import strings

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lshigami/Plog/docs" // Import the docs package to register Swagger docs
	"github.com/lshigami/Plog/internal/config"
	"github.com/lshigami/Plog/internal/db/sqlc"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Constants for static file paths (nên định nghĩa một chỗ)
const (
	staticRootPath  = "/app/static" // Thư mục chứa file static build trong container
	staticIndexFile = "index.html"  // Tên file HTML chính
	// staticAssetsPath = "/app/static/assets" // Đường dẫn tới thư mục assets nếu có cấu trúc riêng
	// urlAssetsPrefix  = "/assets"          // URL prefix cho assets
)

func SetupRouter(store sqlc.Querier, cfg config.Config) *gin.Engine {

	router := gin.Default() // logger & recovery middleware

	// --- CORS Configuration ---
	// (Giữ nguyên hoặc điều chỉnh nếu cần cho cả API và frontend)
	corsConfig := cors.DefaultConfig()
	// Cho phép origin của frontend dev và có thể cả domain production
	corsConfig.AllowOrigins = []string{"http://localhost:3000"} // Giả sử có ClientOrigin trong config
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowCredentials = true // Quan trọng nếu có dùng cookie/session với frontend
	router.Use(cors.New(corsConfig))

	// --- Server Instance ---
	// Giả sử NewServer giờ nhận cfg và store (đã đúng theo code bạn cung cấp)
	server := NewServer(cfg, store)
	server.router = router // Gán router cho server instance

	// --- API Routes ---
	// Nhóm tất cả API dưới một prefix chung (ví dụ: /api/v1)
	apiV1 := router.Group("/api/v1")
	{
		// Swagger documentation endpoint (đặt dưới API prefix)
		apiV1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		// Public API routes
		apiV1.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
		apiV1.POST("/register", server.RegisterUser)
		apiV1.POST("/login", server.LoginUser)

		// Post API routes (Public GET)
		postRoutes := apiV1.Group("/posts")
		{
			postRoutes.GET("", server.ListPosts)
			postRoutes.GET("/:id", server.GetPost)
		}

		// Authenticated API routes (sử dụng AuthMiddleware)
		// authRoutes cần được nhóm bên trong apiV1 để có prefix /api/v1
		authRoutes := apiV1.Group("/") // Áp dụng middleware cho các route bên dưới trong group này
		// Giả sử AuthMiddleware nhận TokenMaker từ server instance
		// Nếu AuthMiddleware không cần TokenMaker, bỏ server.tokenMaker đi
		authRoutes.Use(AuthMiddleware(server.tokenMaker))
		{
			// POST /api/v1/posts
			authRoutes.POST("/posts", server.CreatePost)
			// PUT /api/v1/posts/:id
			authRoutes.PUT("/posts/:id", server.UpdatePost)
			// Thêm các authenticated routes khác ở đây nếu cần
			// Ví dụ: DELETE /api/v1/posts/:id
			// authRoutes.DELETE("/posts/:id", server.DeletePost)
		}
	} // Kết thúc group /api/v1

	// --- Static Frontend Files Serving ---
	// Serve các file static từ thư mục đã định nghĩa (staticRootPath)

	// 1. Phục vụ các file cụ thể như manifest, logo (nếu cần truy cập trực tiếp)
	// router.StaticFileFS("/manifest.json", filepath.Join(staticRootPath, "manifest.json"), http.Dir(staticRootPath))
	// router.StaticFileFS("/logo192.png", filepath.Join(staticRootPath, "logo192.png"), http.Dir(staticRootPath))

	// 2. Phục vụ thư mục chứa assets (CSS, JS, images, fonts)
	// URL path bắt đầu bằng /static/ hoặc /assets/ sẽ tìm file trong staticRootPath
	// Ví dụ: /static/css/main.css -> /app/static/static/css/main.css (điều chỉnh nếu cần)
	// Hoặc nếu build output có thư mục 'assets':
	// router.StaticFS(urlAssetsPrefix, http.Dir(staticAssetsPath))
	router.StaticFS("/static", http.Dir(staticRootPath+"/static")) // Phục vụ thư mục /app/static/static qua URL /static
	router.StaticFS("/assets", http.Dir(staticRootPath+"/assets")) // Phục vụ thư mục /app/static/assets qua URL /assets (Nếu có)

	// 3. Phục vụ file index.html cho đường dẫn gốc "/"
	router.StaticFileFS("/", filepath.Join(staticRootPath, staticIndexFile), http.Dir(staticRootPath))

	// --- SPA Catch-all Route ---
	router.NoRoute(func(c *gin.Context) {
		// Chỉ xử lý cho phương thức GET và không phải là API hoặc file static đã biết
		if c.Request.Method == http.MethodGet &&
			!strings.HasPrefix(c.Request.URL.Path, "/api/") && // Không phải API
			!strings.HasPrefix(c.Request.URL.Path, "/static/") && // Không phải static assets
			!strings.HasPrefix(c.Request.URL.Path, "/assets/") && // Không phải assets (nếu có)
			!strings.HasPrefix(c.Request.URL.Path, "/swagger/") { // Không phải swagger

			// Kiểm tra xem file có tồn tại vật lý trong thư mục static không
			// Điều này tránh trả về index.html cho các lỗi 404 của file thật (ví dụ: logo.png bị thiếu)
			filePath := filepath.Join(staticRootPath, filepath.Clean(c.Request.URL.Path))
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				// File không tồn tại vật lý -> khả năng cao là route của SPA
				// Trả về file index.html chính
				c.File(filepath.Join(staticRootPath, staticIndexFile))
				return
			}
			// Nếu file tồn tại nhưng không được serve bởi StaticFS/StaticFileFS ở trên
			// (ví dụ: cấu hình thiếu), thì để Gin xử lý 404 mặc định
		}

		// Nếu không phải trường hợp SPA hoặc là API/static không tìm thấy, trả 404 mặc định
		// Hoặc bạn có thể trả JSON 404 như trước
		// c.JSON(http.StatusNotFound, HTTPError{Error: "Resource not found"})
	})

	return router
}
