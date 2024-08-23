package api

import (
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/jackc/pgx/v5"
)

// getAccounts godoc
//
//	@Summary	List all budgeting accounts
//	@Schemes
//	@Description	List all accounts associated with a budget.
//	@Param			budget_id	path	string	true	"Budget ID"
//	@Tags			Accounts
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	[]db.CreateAccountParams
//	@Failure		400	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/accounts [get]
func (s *Server) getAccounts(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	accounts, err := s.db.GetAccounts(ctx, budgetId)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}

// getAccount godoc
//
//	@Summary	Get a single budgeting account
//	@Schemes
//	@Description	Get the details of an account associated with a budget.
//	@Param			budget_id	path	string	true	"Budget ID"
//	@Param			account_id	path	string	true	"Account ID"
//	@Tags			Accounts
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	db.CreateAccountParams
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/accounts/{account_id} [get]
func (s *Server) getAccount(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	var acctRqst AccountId
	if err := ctx.ShouldBindUri(&acctRqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}
	// convert to UUID
	acctId, err := uuid.Parse(acctRqst.AccountId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// Get the account
	account, err := s.db.GetAccount(ctx, db.GetAccountParams{
		BudgetID: budgetId,
		ID:       acctId,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse("account not found in budget"))
			return
		}
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

// createAccount godoc
//
//	@Summary	Create a budgeting account
//	@Schemes
//	@Description	Create a budgeting account.
//	@Param			budget_id	path	string					true	"Budget ID"
//	@Param			account		body	updateAccountRequest	true	"Account details"
//	@Tags			Accounts
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	db.CreateAccountParams
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/accounts [post]
func (s *Server) createAccount(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	var rqst accountRequest
	if err := ctx.ShouldBindJSON(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// TODO: improve validation of account type
	accountTypes := []string{"savings", "checking", "lineofcredit"}
	if !slices.Contains(accountTypes, strings.ToLower(rqst.Type)) {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid account type"))
		return
	}

	// Create the account
	// TODO: this should be in a DB transaction, which includes createing a transaction for the starting balance
	arg := db.CreateAccountParams{
		BudgetID: budgetId,
		Name:     rqst.Name,
		Type:     rqst.Type,
		Balance:  rqst.Balance,
	}
	account, err := s.db.CreateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

// updateAccount godoc
//
//	@Summary	Update a budgeting account
//	@Schemes
//	@Description	Update a budgeting account.
//	@Param			budget_id	path	string					true	"Budget ID"
//	@Param			account_id	path	string					true	"Account ID"
//	@Param			account		body	updateAccountRequest	true	"Account details"
//	@Tags			Accounts
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	db.UpdateAccountParams
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/accounts/{account_id} [put]
func (s *Server) updateAccount(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	var acctRqst AccountId
	if err := ctx.ShouldBindUri(&acctRqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}
	// convert to UUID
	acctId, err := uuid.Parse(acctRqst.AccountId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// Take in the request
	var rqst updateAccountRequest
	if err := ctx.ShouldBindJSON(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// Validations
	if rqst.Type.Valid {
		accountTypes := []string{"savings", "checking", "lineofcredit"}
		if !slices.Contains(accountTypes, strings.ToLower(rqst.Type.String)) {
			ctx.JSON(http.StatusBadRequest, errorResponse("invalid account type"))
			return
		}
	}

	// Send the update
	arg := db.UpdateAccountParams{
		ID:               acctId,
		BudgetID:         budgetId,
		Name:             rqst.Name,
		Type:             rqst.Type,
		Closed:           rqst.Closed,
		Note:             rqst.Note,
		Balance:          rqst.Balance,
		ClearedBalance:   rqst.ClearedBalance,
		UnclearedBalance: rqst.UnclearedBalance,
	}
	updatedAccount, err := s.db.UpdateAccount(ctx, arg)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, updatedAccount)
}

// deleteAccount godoc
//
//	@Summary	Delete a budgeting account
//	@Schemes
//	@Description	Delete a budgeting account.
//	@Param			budget_id	path	string	true	"Budget ID"
//	@Param			account_id	path	string	true	"Account ID"
//	@Tags			Accounts
//	@Produce		json
//	@Success		200	{object}	string	"budgeting account deleted"
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/accounts/{account_id} [delete]
func (s *Server) deleteAccount(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	var acctRqst AccountId
	if err := ctx.ShouldBindUri(&acctRqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}
	// convert to UUID
	acctId, err := uuid.Parse(acctRqst.AccountId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// Verify that the account is reconciled and balance is 0
	acct, err := s.db.GetAccount(ctx, db.GetAccountParams{
		ID:       acctId,
		BudgetID: budgetId,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse("budget account not found or no permission"))
			return
		}
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}
	if acct.Balance != 0 || acct.ClearedBalance != 0 {
		ctx.JSON(http.StatusBadRequest, errorResponse("account balance or reconciled balance is not 0"))
		return
	}

	// Delete the account
	// TODO: this should be in a DB transaction, which includes deleting transactions associated with the account
	err = s.db.DeleteAccount(ctx, acctId)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "budgeting account deleted"})
}
