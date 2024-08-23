package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/jackc/pgx/v5"
)

// getCategoryGroups godoc
//
//	@Summary	List all budgeting category groups
//	@Schemes
//	@Description	List all category groups associated with a budget.
//	@Param			budget_id	path	string	true	"Budget ID"
//	@Tags			Categories
//	@Produce		json
//	@Success		200	{object}	[]db.CategoryGroup
//	@Failure		400	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/category-groups [get]
func (s *Server) getCategoryGroups(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	cgs, err := s.db.GetCategoryGroupsByBudgetId(ctx, budgetId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, cgs)
}

// createCategoryGroup godoc
//
//	@Summary	Create a budgeting category group
//	@Schemes
//	@Description	Create a budgeting category group.
//	@Param			budget_id	path	string				true	"Budget ID"
//	@Param			account		body	categoryGroupRqst	true	"Category group details"
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	db.CategoryGroup
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/category-groups [post]
func (s *Server) createCategoryGroup(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	var rqst categoryGroupRqst
	if err := ctx.ShouldBindJSON(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	resp, err := s.db.CreateCategoryGroup(ctx, db.CreateCategoryGroupParams{
		BudgetID: budgetId,
		Name:     rqst.Name,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// updateCategoryGroup godoc
//
//	@Summary	Update a budgeting category group
//	@Schemes
//	@Description	Update a budgeting category group.
//	@Param			budget_id			path	string				true	"Budget ID"
//	@Param			category_group_id	path	string				true	"Category Group ID"
//	@Param			account				body	categoryGroupRqst	true	"Category group details"
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	db.CategoryGroup
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/category-groups/{category_group_id} [put]
func (s *Server) updateCategoryGroup(ctx *gin.Context) {

	// Parse the GET parameters
	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}
	var cGroupId CategoryGroupId
	if err := ctx.ShouldBindUri(&cGroupId); err != nil {
		slog.Info(err.Error())
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}
	cGroupUuid, err := uuid.Parse(cGroupId.CategoryGroupId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// Parse the JSON request body
	var rqst categoryGroupRqst
	if err := ctx.ShouldBindJSON(&rqst); err != nil {
		slog.Info("Invalid body")
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// Update
	newcGroup, err := s.db.UpdateCategoryGroup(ctx, db.UpdateCategoryGroupParams{
		ID:   cGroupUuid,
		Name: rqst.Name,
	})
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, newcGroup)
}

// deleteCategoryGroup godoc
//
//	@Summary	Delete a category group
//	@Schemes
//	@Description	Delete a budgeting category group and all categories in the group.
//	@Param			budget_id			path	string	true	"Budget ID"
//	@Param			category_group_id	path	string	true	"Category Group ID"
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	string	"category group deleted"
//	@Failure		400	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/category-groups/{category_group_id} [delete]
func (s *Server) deleteCategoryGroup(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	var categoryGroupId CategoryGroupId
	if err := ctx.ShouldBindUri(&categoryGroupId); err != nil {
		slog.Info(err.Error())
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}
	categoryGroupUuid, err := uuid.Parse(categoryGroupId.CategoryGroupId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	_, err = s.db.GetCategoryGroup(ctx, categoryGroupUuid)
	if err != nil && err == pgx.ErrNoRows {
		ctx.JSON(http.StatusNotFound, errorResponse("category group id not found"))
		return
	}

	// TODO: this should be in a DB transaction, which includes deleting transactions associated with the category group

	// Call the transaction to delete the category group
	err = s.db.DeleteCategoryGroupTx(ctx, categoryGroupUuid)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "category group deleted"})
}
