package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/exp/slog"
)

// GET /verify_email?id=XX&code=YYYY
func (s *Server) verifyEmail(ctx *gin.Context) {

	var rqst verifyEmailRequest
	if err := ctx.ShouldBindQuery(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// retrieve from the DB
	e, err := s.db.GetVerifyEmails(ctx, db.GetVerifyEmailsParams{
		ID:   rqst.ID,
		Code: rqst.Code,
	})
	if err != nil {
		if err == pgx.ErrNoRows || err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, errorResponse("invalid code"))
			return
		}
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	// retrieve the user from the database
	u, err := s.db.GetUserByUsername(ctx, e.Username)
	if err != nil {
		if err == pgx.ErrNoRows || err == sql.ErrNoRows {
			// the user was deleted from the time they signed up to clicking the verify email link
			ctx.JSON(http.StatusBadRequest, errorResponse("invalid user"))
			return
		}
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	// Validations
	if u.EmailVerified {
		ctx.JSON(http.StatusBadRequest, errorResponse("email is already verified"))
		return
	}
	if time.Now().After(e.ExpiresAt) {
		ctx.JSON(http.StatusBadRequest, errorResponse("code is expired, please login again to resend verification email"))
		return
	}
	// edge case that the code was used but still not verified?
	if e.Used {
		ctx.JSON(http.StatusBadRequest, errorResponse("code was already used"))
		return
	}

	// set used to true and set user emailverified to true
	arg := db.UpdateUserParams{
		Username: u.Username,
		EmailVerified: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
	}
	updatedUser, err := s.db.UpdateUser(ctx, arg)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}
	_, err = s.db.UpdateCodeUsed(ctx, e.Code)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	resp := userResponse{
		Username:      updatedUser.Username,
		Email:         updatedUser.Email,
		EmailVerified: updatedUser.EmailVerified,
	}
	ctx.JSON(http.StatusOK, resp)
}
