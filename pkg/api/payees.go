package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/jackc/pgx/v5"
)

// getPayees godoc
//
//	@Summary	List payees
//	@Schemes
//	@Description	Get all payees
//	@Param			budget_id	path	string	true	"Budget ID"
//	@Tags			Payees
//	@Produce		json
//	@Success		200	{object}	[]db.Payee
//	@Failure		400	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/payees [get]
func (s *Server) getPayees(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	payees, err := s.db.GetPayees(ctx, budgetId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, payees)
}

// getPayee godoc
//
//	@Summary	Get payee
//	@Schemes
//	@Description	Get a payee by id
//	@Param			budget_id	path	string	true	"Budget ID"
//	@Param			payee_id	path	string	true	"Payee ID"
//	@Tags			Payees
//	@Produce		json
//	@Success		200	{object}	db.Payee
//	@Failure		400	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/payees/{payee_id} [get]
func (s *Server) getPayee(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	// Get the payee
	var rqst PayeeId
	if err := ctx.ShouldBindUri(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}
	payeeId, err := uuid.Parse(rqst.PayeeId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	payee, err := s.db.GetPayeeById(ctx, payeeId)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse("payee not found"))
			return
		}
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, payee)
}

// createPayee godoc
//
//	@Summary	Create a payee
//	@Schemes
//	@Description	Create a spending recipient.
//	@Param			budget_id	path	string		true	"Budget ID"
//	@Param			account		body	payeeRqst	true	"Payee details"
//	@Tags			Payees
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	db.Payee
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/payees [post]
func (s *Server) createPayee(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	var rqst payeeRqst
	if err := ctx.ShouldBindJSON(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	resp, err := s.db.CreatePayee(ctx, db.CreatePayeeParams{
		BudgetID: budgetId,
		Name:     rqst.Name,
	})
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// updatePayee godoc
//
//	@Summary	Update a payee
//	@Schemes
//	@Description	Update a payee.
//	@Param			budget_id	path	string		true	"Budget ID"
//	@Param			Payee_id	path	string		true	"Payee ID"
//	@Param			account		body	payeeRqst	true	"Payee details"
//	@Tags			Payees
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	db.Payee
//	@Failure		400	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/payees/{payee_id} [put]
func (s *Server) updatePayee(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	// Get the payee details
	var payeeId PayeeId
	if err := ctx.ShouldBindUri(&payeeId); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}
	payeeUuid, err := uuid.Parse(payeeId.PayeeId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// Parse the JSON request body
	var rqst payeeRqst
	if err := ctx.ShouldBindJSON(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// Update
	newPayee, err := s.db.UpdatePayee(ctx, db.UpdatePayeeParams{
		Name:     rqst.Name,
		BudgetID: budgetId,
		ID:       payeeUuid,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse("payee not found"))
			return
		}
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, newPayee)
}

// deletePayee godoc
//
//	@Summary	Delete a payee
//	@Schemes
//	@Description	Delete a payee.
//	@Param			budget_id	path	string	true	"Budget ID"
//	@Param			Payee_id	path	string	true	"Payee ID"
//	@Tags			Payees
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	string	"payee deleted"
//	@Failure		400	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/payees/{payee_id} [delete]
func (s *Server) deletePayee(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	// Get the payee details
	var payeeId PayeeId
	if err := ctx.ShouldBindUri(&payeeId); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}
	payeeUuid, err := uuid.Parse(payeeId.PayeeId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// TODO: this should be in a DB transaction, which includes deleting transactions associated with the payee

	err = s.db.DeletePayee(ctx, db.DeletePayeeParams{
		BudgetID: budgetId,
		ID:       payeeUuid,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse("payee not found"))
			return
		}
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "payee deleted"})
}
