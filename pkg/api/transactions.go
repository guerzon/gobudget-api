package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// getTransactions godoc
//
//	@Summary	List all transactions
//	@Schemes
//	@Description	List all transactions across all accounts in the budget.
//	@Param			budget_id	path	string	true	"Budget ID"
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	[]db.TransactionsView
//	@Failure		400	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/transactions [get]
func (s *Server) getTransactions(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	// Get the transactions
	transactions, err := s.db.GetTransactionsView(ctx, budgetId)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, transactions)
}

// getTransaction godoc
//
//	@Summary	Get a transaction
//	@Schemes
//	@Description	Get the details of a transaction
//	@Param			budget_id		path	string	true	"Budget ID"
//	@Param			transaction_id	path	string	true	"Transaction ID"
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	transactionResponse
//	@Failure		400	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/transactions/{transaction_id} [get]
func (s *Server) getTransaction(ctx *gin.Context) {

	// Parse the IDs
	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}
	var rqst TransactionId
	if err := ctx.ShouldBindUri(&rqst); err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request 1"))
		return
	}
	transactionId, err := uuid.Parse(rqst.Id)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// Get the transaction
	transaction, err := s.db.GetTransactionsViewById(ctx, transactionId)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	// Build the response
	resp := transactionResponse{
		Date:       transaction.Date,
		Account:    transaction.AccountName,
		Payee:      transaction.PayeeName,
		Category:   transaction.CategoryName,
		Memo:       transaction.Memo,
		Amount:     transaction.Amount,
		Approved:   transaction.Approved,
		Cleared:    transaction.Cleared,
		Reconciled: transaction.Reconciled,
	}
	ctx.JSON(http.StatusOK, resp)
}

// createTransaction godoc
//
//	@Summary	Create a transaction
//	@Schemes
//	@Description	Create a transaction.
//	@Param			budget_id	path	string				true	"Budget ID"
//	@Param			transaction	body	transactionRequest	true	"Transaction details"
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	db.TransactionsView
//	@Failure		400	{object}	HTTPError
//	@Failure		403	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/transactions [post]
func (s *Server) createTransaction(ctx *gin.Context) {

	// Parse the request
	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		ctx.JSON(http.StatusForbidden, errorResponse("budget does not exist or does not belong to the user"))
		return
	}
	var rqst transactionRequest
	if err := ctx.ShouldBindJSON(&rqst); err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"+err.Error()))
		return
	}

	// Make sure that the account in the POST body belongs to the user
	accountId, err := uuid.Parse(rqst.Account)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusBadRequest, errorResponse("cannot parse UUID: "+err.Error()))
		return
	}
	acct, err := s.db.GetAccount(ctx, db.GetAccountParams{
		BudgetID: budgetId,
		ID:       accountId,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			slog.Error(err.Error())
			ctx.JSON(http.StatusForbidden, errorResponse("account does not exist or does not belong to the user"))
			return
		}
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	// Validations
	if rqst.Account == rqst.Payee {
		ctx.JSON(http.StatusBadRequest, errorResponse("account and payee should not be the same"))
		return
	}

	// Parse the UUIDs
	payeeId, err := uuid.Parse(rqst.Payee)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("cannot parse payee ID"))
		return
	}
	categoryId, err := uuid.Parse(rqst.Category)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("cannot parse category ID"))
		return
	}

	// Create the transaction
	// TODO: this should be in a DB transaction along with updating the balance of the account
	// And if the payee is also an account, it should also be updated
	arg := db.CreateTransactionParams{
		AccountID: acct.ID,
		Date: pgtype.Date{
			Valid: true,
			Time:  rqst.Date.Time,
		},
		PayeeID: payeeId,
		CategoryID: pgtype.UUID{
			Bytes: categoryId,
			Valid: true,
		},
		Memo: pgtype.Text{
			Valid:  true,
			String: rqst.Memo,
		},
		Amount:     rqst.Amount,
		Cleared:    rqst.Cleared,
		Reconciled: rqst.Reconciled,
	}
	resp, err := s.db.CreateTransaction(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				ctx.JSON(http.StatusBadRequest, errorResponse("invalid payee or category ID"))
				return
			}
		}
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	// Debit the account's balance

	// Check if the payee is an account and debit as well

	ctx.JSON(http.StatusOK, resp)
}

// updateTransaction godoc
//
//	@Summary	Update a transaction
//	@Schemes
//	@Description	Update a transaction.
//	@Param			budget_id		path	string				true	"Budget ID"
//	@Param			transaction_id	path	string				true	"Transaction ID"
//	@Param			transaction		body	transactionRequest	true	"Transaction details"
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	db.TransactionsView
//	@Failure		400	{object}	HTTPError
//	@Failure		403	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/transactions/:transaction_id [put]
// func (s *Server) updateTransaction(ctx *gin.Context) {

// 	// Parse the request
// 	var budgetId uuid.UUID
// 	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
// 		return
// 	}
// 	var transactionUri TransactionId
// 	transactionId, err := uuid.Parse(transactionUri.Id)
// 	if err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
// 		return
// 	}
// 	var rqst transactionRequest
// 	if err := ctx.ShouldBindJSON(&rqst); err != nil {
// 		slog.Error(err.Error())
// 		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
// 		return
// 	}

// 	// Make sure that the account in the POST body belongs to the user
// 	acct, err := s.db.GetAccount(ctx, db.GetAccountParams{
// 		BudgetID: budgetId,
// 		ID:       rqst.Account,
// 	})
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			ctx.JSON(http.StatusForbidden, errorResponse("account does not exist or does not belong to the user"))
// 			return
// 		}
// 		slog.Error(err.Error())
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
// 		return
// 	}

// 	// Validations
// 	if rqst.Account == rqst.Payee {
// 		ctx.JSON(http.StatusBadRequest, errorResponse("account and payee should not be the same"))
// 		return
// 	}

// 	// send the update
// 	arg := db.UpdateTransactionParams{
// 		ID:        transactionId,
// 		AccountID: rqst.Account,
// 	}
// }
