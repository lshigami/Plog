package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lshigami/Plog/internal/auth"
)

const (
	AuthorizationHeaderKey  = "Authorization"
	AuthorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = "authorization_payload"
	UserIDKey               = "user_id"
)

func AuthMiddleware(maker auth.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		authorizationHeader := ctx.GetHeader(AuthorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}
		fields := strings.Fields(authorizationHeader)

		if len(fields) < 2 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
			return
		}

		authType := strings.ToLower(fields[0])
		if authType != strings.ToLower(AuthorizationTypeBearer) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unsupported authorization type: " + authType})
			return
		}

		accessToken := fields[1]
		claims, err := maker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx.Set(AuthorizationPayloadKey, claims)
		ctx.Set(UserIDKey, claims.ID)
		ctx.Next()

	}
}
