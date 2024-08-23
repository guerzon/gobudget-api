package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/guerzon/gobudget-api/pkg/token"
	"github.com/guerzon/gobudget-api/pkg/util"
	"github.com/guerzon/gobudget-api/pkg/worker"
	"github.com/jackc/pgx/v5/pgconn"
)

// createUser godoc
//
//	@Summary	Create user
//	@Schemes
//	@Description	Create a new user account. An email is sent asking the user to verify their email.
//	@Tags			User
//	@Accept			json
//	@Param			account	body	createUserRequest	true	"Create a new user"
//	@Produce		json
//	@Success		201	{object}	userResponse
//	@Failure		400	{object}	HTTPError
//	@Failure		401	{object}	HTTPError
//	@Failure		403	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/user [post]
func (s *Server) createUser(ctx *gin.Context) {

	var rqst createUserRequest

	if err := ctx.ShouldBindJSON(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	hashedPassword, err := util.HashPassword(rqst.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	arg := db.CreateUserParams{
		Username:           rqst.Username,
		Password:           hashedPassword,
		Email:              rqst.Email,
		LastPasswordChange: time.Now(),
	}

	// Call the transaction to create the user
	afterCreateFn := func(createdUser db.UserParams) error {
		taskPayload := &worker.SendEmailPayload{
			Username: createdUser.Username,
		}
		return s.taskDistributor.DistributeSendEmail(ctx, taskPayload, worker.TaskSendVerifyEmail)
	}
	user, err := s.db.CreateUserTx(ctx, arg, afterCreateFn)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				ctx.JSON(http.StatusForbidden, errorResponse("user already exists"))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	slog.Info("Created new user", "user", user.Username)

	resp := userResponse{
		Username:      user.Username,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}

	ctx.JSON(http.StatusCreated, resp)
}

// updateUser godoc
//
//	@Summary	Update user
//	@Schemes
//	@Description	Update the authenticated user's account.
//	@Tags			User
//	@Param			account	body	updateUserRequest	true	"Update account"
//	@Produce		json
//	@Success		200	{object}	userResponse
//	@Failure		400	{object}	HTTPError
//	@Failure		401	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/user [put]
//	@Security		Bearer
func (s *Server) updateUser(ctx *gin.Context) {

	// Get the authenticated user
	k, exists := ctx.Get("authz_payload")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}
	authz_payload := k.(*token.TokenPayload)

	// Get the user from the database
	userBefore, err := s.db.GetUserByUsername(ctx, authz_payload.Username)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse("user is not valid"))
		return
	}

	var rqst updateUserRequest
	if err := ctx.ShouldBindJSON(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("cannot update user, invalid request"))
		return
	}
	arg := db.UpdateUserParams{
		Email:    rqst.Email,
		Password: rqst.Password,
		Username: userBefore.Username,
	}

	if arg.Email.Valid && userBefore.Email != arg.Email.String {
		slog.Info("Update EmailVerified")
		arg.EmailVerified.Bool = false
		arg.EmailVerified.Valid = true
	}
	if arg.Password.Valid {
		if err := util.CheckPassword(userBefore.Password, arg.Password.String); err != nil {
			hashedPassword, err := util.HashPassword(rqst.Password.String)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
				return
			}
			arg.Password.String = hashedPassword
			arg.LastPasswordChange.Valid = true
			arg.LastPasswordChange.Time = time.Now()
		}
	}

	fmt.Printf("Updating LastPasswordChange to: %v\n", arg.LastPasswordChange.Time)

	// Call the transaction to update the user
	afterUpdateFn := func(createdUser db.UserParams) error {
		taskPayload := &worker.SendEmailPayload{
			Username: createdUser.Username,
		}
		return s.taskDistributor.DistributeSendEmail(ctx, taskPayload, worker.TaskSendVerifyEmail)
	}
	updatedUser, err := s.db.UpdateUserTx(ctx, arg, afterUpdateFn)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	resp := userResponse{
		Username:           updatedUser.Username,
		Email:              updatedUser.Email,
		EmailVerified:      updatedUser.EmailVerified,
		CreatedAt:          updatedUser.CreatedAt,
		LastPasswordChange: updatedUser.LastPasswordChange,
	}

	ctx.JSON(http.StatusOK, resp)
}

// deleteUser godoc
//
//	@Summary	Delete user
//	@Schemes
//	@Description	Delete the authenticated user's account.
//	@Tags			User
//	@Produce		json
//	@Success		200	{string}	string	"user has been deleted"
//	@Failure		401	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/user [delete]
//	@Security		Bearer
func (s *Server) deleteUser(ctx *gin.Context) {

	// Get the authenticated user
	k, exists := ctx.Get("authz_payload")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}
	authz_payload := k.(*token.TokenPayload)

	// Get the user from the database
	user, err := s.db.GetUserByUsername(ctx, authz_payload.Username)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse("user is not valid"))
		return
	}

	// Get the user's budget accounts
	budgets, err := s.db.GetBudgets(ctx, user.Username)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}
	budgetIds := make([]uuid.UUID, len(budgets))
	for b := range budgets {
		budgetIds[b] = budgets[b].ID
		fmt.Println(budgetIds[b].String())
	}

	// Call the db transaction to delete a user
	afterDeleteFn := func(deletedUser db.UserParams) error {
		payload := &worker.SendEmailPayload{
			Username: deletedUser.Username,
			Email:    deletedUser.Email,
		}
		return s.taskDistributor.DistributeSendEmail(ctx, payload, worker.TaskSendAccountDeletedEmail)
	}
	userArg := db.UserParams{
		Username: user.Username,
		Email:    user.Email,
	}
	err = s.db.DeleteUserTx(ctx, userArg, budgetIds, afterDeleteFn)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "user has been deleted"})
}
