package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/guerzon/gobudget-api/pkg/token"
	"github.com/jackc/pgx/v5"
)

// getBudgets godoc
//
//	@Summary	List budgets
//	@Schemes
//	@Description	List all budgets.
//	@Tags			Budget
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	[]db.CreateBudgetParams
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets [get]
func (s *Server) getBudgets(ctx *gin.Context) {

	// Get the authenticated user
	k, exists := ctx.Get("authz_payload")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}
	authz_payload := k.(*token.TokenPayload)

	resp, err := s.db.GetBudgets(ctx, authz_payload.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// getBudget godoc
//
//	@Summary	Get budget
//	@Schemes
//	@Description	Get the details of a budget.
//	@Tags			Budget
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	detailedBudgetResponse
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/:budget_id [get]
func (s *Server) getBudget(ctx *gin.Context) {

	// Get the authenticated user
	k, exists := ctx.Get("authz_payload")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}
	authz_payload := k.(*token.TokenPayload)

	var budgetRqst BudgetId
	if err := ctx.ShouldBindUri(&budgetRqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("valid budget ID not found"))
		return
	}
	// convert to UUID
	budgetId, err := uuid.Parse(budgetRqst.BudgetId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("valid budget ID not found"))
		return
	}

	budget, err := s.db.GetBudget(ctx, db.GetBudgetParams{
		OwnerUsername: authz_payload.Username,
		ID:            budgetId,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse("this budget does not exist or user has no permission"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	// get the associated accounts
	accounts, err := s.db.GetAccounts(ctx, budgetId)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	resp := detailedBudgetResponse{
		Id:           budget.ID,
		Name:         budget.Name,
		CurrencyCode: budget.CurrencyCode,
		Accounts:     accounts,
	}

	ctx.JSON(http.StatusOK, resp)
}

// createBudget godoc
//
//	@Summary	Create budget
//	@Schemes
//	@Description	Create a new budget.
//	@Tags			Budget
//	@Accept			json
//	@Param			account	body	db.CreateBudgetParams true	"Create a budget"
//	@Produce		json
//	@Success		200	{object}	db.CreateBudgetParams
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets [post]
func (s *Server) createBudget(ctx *gin.Context) {

	// Get the authenticated user
	k, exists := ctx.Get("authz_payload")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}
	authz_payload := k.(*token.TokenPayload)

	var rqst budgetRequest
	if err := ctx.ShouldBindJSON(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	arg := db.CreateBudgetParams{
		OwnerUsername: authz_payload.Username,
		Name:          rqst.Name,
		CurrencyCode:  rqst.CurrencyCode,
	}

	// Check if the budget already exists
	budget, err := s.db.GetBudgetDetails(ctx, db.GetBudgetDetailsParams{
		OwnerUsername: authz_payload.Username,
		Name:          rqst.Name,
		CurrencyCode:  rqst.CurrencyCode,
	})
	if err != nil && err != pgx.ErrNoRows {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}
	if budget.Name != "" {
		ctx.JSON(http.StatusBadRequest, errorResponse("budget "+rqst.Name+" already exists"))
		return
	}

	// TODO: validate currency

	resp, err := s.db.CreateBudget(ctx, arg)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// BUG: this will delete all budgets with the same name

// deleteBudget godoc
//
//	@Summary	Delete budget
//	@Schemes
//	@Description	Delete a budget.
//	@Tags			Budget
//	@Param			budget_id	path	string	true	"Budget ID"
//	@Produce		json
//	@Success		200	{object}	string "budget deleted"
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets [delete]
func (s *Server) deleteBudget(ctx *gin.Context) {

	// Get the authenticated user
	k, exists := ctx.Get("authz_payload")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}
	authz_payload := k.(*token.TokenPayload)

	var budgetRqst BudgetId
	if err := ctx.ShouldBindUri(&budgetRqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("valid budget ID not found"))
		return
	}
	// convert to UUID
	budgetId, err := uuid.Parse(budgetRqst.BudgetId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("valid budget ID not found"))
		return
	}

	// Verify that the budget exists and the user owns it
	_, err = s.db.GetBudget(ctx, db.GetBudgetParams{
		ID:            budgetId,
		OwnerUsername: authz_payload.Username,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse("this budget does not exist or user has no permission"))
			return
		}
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	// Call the transaction to delete the budget
	err = s.db.DeleteBudgetTx(ctx, budgetId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		slog.Error(err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "budget deleted"})
}
