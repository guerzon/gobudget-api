package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/jackc/pgx/v5"
)

// getCategories godoc
//
//	@Summary	Get category groups and categories
//	@Schemes
//	@Description	List all categories in a budget grouped by category group
//	@Param			budget_id	path	string	true	"Budget ID"
//	@Tags			Categories
//	@Produce		json
//	@Success		200	{object}	categoryResponse
//	@Failure		400	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/categories [get]
func (s *Server) getCategories(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	categoryGroups, err := s.db.GetCategoryGroupsByBudgetId(ctx, budgetId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	resp := make([]categoryResponse, len(categoryGroups))

	fmt.Printf("categoryGroups length is: %d\n", len(categoryGroups))
	// Get the category group and the categories
	for c := range categoryGroups {
		cgrp, err := s.db.GetCategoryGroup(ctx, categoryGroups[c].ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
			return
		}
		resp[c].CategoryGroupId = cgrp.ID
		resp[c].Name = cgrp.Name

		cgs, err := s.db.GetCategories(ctx, cgrp.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
			return
		}
		resp[c].Categories = cgs
	}

	ctx.JSON(http.StatusOK, resp)
}

// createCategory godoc
//
//	@Summary	Create a budgeting category
//	@Schemes
//	@Description	Create a budgeting category.
//	@Param			budget_id			path	string			true	"Budget ID"
//	@Param			category_group_id	path	string			true	"Category Group ID"
//	@Param			account				body	categoryRqst	true	"Category details"
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	db.Category
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/categories/{category_group_id} [post]
func (s *Server) createCategory(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	var rqst categoryRqst
	var rqst2 CategoryGroupId
	if err := ctx.ShouldBindJSON(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}
	if err := ctx.ShouldBindUri(&rqst2); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}
	categoryGroupId, err := uuid.Parse(rqst2.CategoryGroupId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	resp, err := s.db.CreateCategory(ctx, db.CreateCategoryParams{
		CategoryGroupID: categoryGroupId,
		Name:            rqst.Name,
	})
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// updateCategory godoc
//
//	@Summary	Update a budgeting category
//	@Schemes
//	@Description	Update a budgeting category.
//	@Param			budget_id	path	string			true	"Budget ID"
//	@Param			category_id	path	string			true	"Category ID"
//	@Param			account		body	categoryRqst	true	"Category details"
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	db.Category
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/categories/{category_id} [put]
func (s *Server) updateCategory(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	var categoryId CategoryId
	if err := ctx.ShouldBindUri(&categoryId); err != nil {
		slog.Info(err.Error())
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}
	categoryUuid, err := uuid.Parse(categoryId.CategoryId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// Parse the JSON request body
	var rqst categoryRqst
	if err := ctx.ShouldBindJSON(&rqst); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// Update
	newcGroup, err := s.db.UpdateCategory(ctx, db.UpdateCategoryParams{
		ID:   categoryUuid,
		Name: rqst.Name,
	})
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, newcGroup)
}

// deleteCategory godoc
//
//	@Summary	Delete a category
//	@Schemes
//	@Description	Delete a category.
//	@Param			budget_id	path	string	true	"Budget ID"
//	@Param			category_id	path	string	true	"Category ID"
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	string	"category deleted"
//	@Failure		400	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/budgets/{budget_id}/categories/{category_id} [delete]
func (s *Server) deleteCategory(ctx *gin.Context) {

	var budgetId uuid.UUID
	if err := s.ValidateBudgetOwnership(ctx, &budgetId); err != nil {
		return
	}

	var categoryId CategoryId
	if err := ctx.ShouldBindUri(&categoryId); err != nil {
		slog.Info(err.Error())
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}
	categoryUuid, err := uuid.Parse(categoryId.CategoryId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse("invalid request"))
		return
	}

	// TODO: this should be in a DB transaction, which includes deleting transactions associated with the category

	err = s.db.DeleteCategory(ctx, categoryUuid)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse("payee not found"))
			return
		}
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, errorResponse(internal_error_message))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "category deleted"})
}
