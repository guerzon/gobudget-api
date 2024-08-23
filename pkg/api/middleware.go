package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/guerzon/gobudget-api/pkg/token"
)

func AuthMiddleware(tokenMaker token.Builder) gin.HandlerFunc {

	return func(ctx *gin.Context) {

		// Get authorization header
		authzHeader := ctx.GetHeader("authorization")
		if len(authzHeader) == 0 {
			e := "authorization header is not provided"
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(e))
			return
		}

		// Parse header
		fields := strings.Fields(authzHeader)
		if len(fields) < 2 {
			e := "invalid authorization header format"
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(e))
			return
		}
		authzType := strings.ToLower(fields[0])
		// only bearer tokens is supported at the moment
		if authzType != "bearer" {
			e := "authorization header is not supported"
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(e))
			return
		}

		// Process the token
		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err.Error()))
			return
		}

		// Hack: prevent the refresh token, which does not have 'aud', from being used for login.
		if len(payload.Audience) == 0 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("invalid session token"))
			return
		}

		// Set authz_payload for next middleware in the chain
		ctx.Set("authz_payload", payload)
		ctx.Next()
	}
}
