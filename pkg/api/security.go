package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/guerzon/gobudget-api/pkg/token"
	"github.com/jackc/pgx/v5"
)

// Validate that the user in the context owns the budget in the URL.
// On success, writes the budgetId to the pointer specified.
func (s *Server) ValidateBudgetOwnership(ctx *gin.Context, budgetId *uuid.UUID) error {

	// Get the authenticated user
	k, exists := ctx.Get("authz_payload")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return errors.New("authz_payload not set")
	}
	authz_payload := k.(*token.TokenPayload)

	var rqst BudgetId
	if err := ctx.ShouldBindUri(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return err
	}
	b, err := uuid.Parse(rqst.BudgetId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return err
	}

	// Ensure the user matches the budget
	_, err = s.db.GetBudget(ctx, db.GetBudgetParams{
		ID:            b,
		OwnerUsername: authz_payload.Username,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse("budget not found or user has no permission"))
			return err
		}
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return err
	}

	*budgetId = b

	return nil
}
