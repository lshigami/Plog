package api

import (
	"github.com/gin-gonic/gin"
	"github.com/lshigami/Plog/internal/auth"
	"github.com/lshigami/Plog/internal/config"
	"github.com/lshigami/Plog/internal/db/sqlc"
)

type Server struct {
	config     config.Config
	store      sqlc.Querier
	tokenMaker auth.Maker
	router     *gin.Engine
}

func NewServer(config config.Config, store sqlc.Querier) *Server {

	tokenMaker := auth.NewJWTMaker(config.JWTSecret)

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}
	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
