package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupSwagger adds Swagger documentation routes to the router
func (server *Server) SetupSwagger(router *gin.Engine) {
	// Use the ginSwagger middleware to serve the API documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
