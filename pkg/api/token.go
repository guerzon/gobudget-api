package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/exp/slog"
)

// renewToken godoc
//
//	@Summary	Renew token
//	@Schemes
//	@Description	Renew access token using a refresh token.
//	@Tags			Security
//	@Accept			json
//	@Param			refresh_token	body	renewTokenRequest	true	"Refresh token"
//	@Produce		json
//	@Success		200	{object}	renewTokenResponse
//	@Failure		401	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/renew_token [post]
func (s *Server) renewToken(ctx *gin.Context) {

	var rqst renewTokenRequest
	if err := ctx.ShouldBindJSON(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// Verify the refresh token in the request
	refreshTokenClaims, err := s.tokenBuilder.VerifyToken(rqst.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err.Error()))
		return
	}

	// Validate the session in the database
	session, err := s.db.GetSession(ctx, refreshTokenClaims.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusUnauthorized, errorResponse("cannot find session with ID "+refreshTokenClaims.ID.String()))
			return
		}
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	// Check if the session is blocked
	if session.IsBlocked {
		ctx.JSON(http.StatusUnauthorized, errorResponse("token is blocked"))
		return
	}

	// These might be unnecessary
	// if session.Username != refreshTokenClaims.Username {
	// 	slog.Error(session.Username)
	// 	ctx.JSON(http.StatusUnauthorized, errorResponse("incorrect user in token"))
	// 	return
	// }
	// if session.RefreshToken != rqst.RefreshToken {
	// 	ctx.JSON(http.StatusUnauthorized, errorResponse("mismatched session token"))
	// 	return
	// }

	// Create access token
	accessToken, accessTokenClaims, err := s.tokenBuilder.CreateToken(refreshTokenClaims.Username, s.config.AccessTokenDuration)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	slog.Info("successfully renewed token for session " + refreshTokenClaims.ID.String())

	resp := renewTokenResponse{
		SessionID:            session.ID,
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessTokenClaims.ExpiresAt.Time,
	}
	ctx.JSON(http.StatusOK, resp)
}
