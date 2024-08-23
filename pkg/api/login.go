package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/guerzon/gobudget-api/pkg/util"
	"github.com/guerzon/gobudget-api/pkg/worker"
	"github.com/jackc/pgx/v5"
)

// POST request to /login containing the username and password
func (s *Server) login(ctx *gin.Context) {

	var rqst loginRequest
	if err := ctx.ShouldBindJSON(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// Validate the email
	u, err := s.db.GetUserByUsername(ctx, rqst.Username)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse("invalid username or password"))
			return
		}
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	// Validate the password
	err = util.CheckPassword(u.Password, rqst.Password)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse("invalid username or password"))
		return
	}

	// Check if the email has been validated
	if !u.EmailVerified {
		// check if there is an active pending record in verify_emails
		resp, err := s.db.GetPendingVerifyEmails(ctx, db.GetPendingVerifyEmailsParams{
			Username:  u.Username,
			Used:      false,
			ExpiresAt: time.Now(),
		})
		if err != nil {
			slog.Error(err.Error())
			ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
			return
		}
		// no valid verification email found, resend
		if len(resp) == 0 {
			taskPayload := &worker.SendEmailPayload{
				Username: u.Username,
			}
			err = s.taskDistributor.DistributeSendEmail(ctx, taskPayload, worker.TaskSendVerifyEmail)
			if err != nil {
				slog.Error(err.Error())
				ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
				return
			}
			ctx.JSON(http.StatusUnauthorized, errorResponse("email not verified, verification email resent"))
			return
		}
		ctx.JSON(http.StatusUnauthorized, errorResponse("email not verified, please check the verification email"))
		return
	}

	// from this point, user is validated

	// create the access token
	accessToken, accessTokenClaims, err := s.tokenBuilder.CreateToken(u.Username, s.config.AccessTokenDuration)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	// create the refresh token
	refreshToken, refreshTokenClaims, err := s.tokenBuilder.CreateToken(u.Username, s.config.RefreshTokenDuration)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	// save the refreshToken in the session
	arg := db.CreateSessionParams{
		ID:           refreshTokenClaims.ID,
		Username:     u.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		ExpiresAt:    refreshTokenClaims.ExpiresAt.Time,
	}
	session, err := s.db.CreateSession(ctx, arg)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	// build and send the response
	resp := loginResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessTokenClaims.ExpiresAt.Time,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenClaims.ExpiresAt.Time,
		ID:                    u.ID,
		Username:              u.Username,
		Email:                 u.Email,
		CreatedAt:             u.CreatedAt,
		LastPasswordChange:    u.LastPasswordChange,
	}

	ctx.JSON(http.StatusOK, resp)
}
