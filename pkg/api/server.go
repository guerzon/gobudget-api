package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	_ "github.com/guerzon/gobudget-api/docs"
	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/guerzon/gobudget-api/pkg/token"
	"github.com/guerzon/gobudget-api/pkg/util"
	"github.com/guerzon/gobudget-api/pkg/worker"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	Router          *gin.Engine
	config          util.Config
	db              db.Store
	tokenBuilder    token.Builder
	taskDistributor worker.TaskDistributor
}

func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {

	// Create a new token builder
	jwtTokenBuilder, err := token.NewJWTBuilder(config.SecretKey)
	if err != nil {
		slog.Error("cannot create a new JWT token builder")
	}

	// create a new server to return
	server := &Server{
		db:              store,
		config:          config,
		tokenBuilder:    jwtTokenBuilder,
		taskDistributor: taskDistributor,
	}

	Router := gin.Default()

	// User facing endpoints, auth required
	beta_users := Router.Group("beta").Use(AuthMiddleware(server.tokenBuilder))
	{
		// User profile actions
		beta_users.PUT("/user", server.updateUser)
		beta_users.DELETE("/user", server.deleteUser)

		// budgets
		beta_users.GET("/budgets", server.getBudgets)
		beta_users.GET("/budgets/:budget_id", server.getBudget)
		beta_users.POST("/budgets", server.createBudget)
		beta_users.DELETE("/budgets/:budget_id", server.deleteBudget)

		// accounts
		beta_users.GET("/budgets/:budget_id/accounts", server.getAccounts)
		beta_users.GET("/budgets/:budget_id/accounts/:account_id", server.getAccount)
		beta_users.POST("/budgets/:budget_id/accounts", server.createAccount)
		beta_users.PUT("/budgets/:budget_id/accounts/:account_id", server.updateAccount)
		beta_users.DELETE("/budgets/:budget_id/accounts/:account_id", server.deleteAccount)

		// category groups
		beta_users.GET("/budgets/:budget_id/category-groups", server.getCategoryGroups)
		beta_users.POST("/budgets/:budget_id/category-groups", server.createCategoryGroup)
		beta_users.PUT("/budgets/:budget_id/category-groups/:category_group_id", server.updateCategoryGroup)
		beta_users.DELETE("/budgets/:budget_id/category-groups/:category_group_id", server.deleteCategoryGroup)

		// categories
		beta_users.GET("/budgets/:budget_id/categories", server.getCategories)
		beta_users.POST("/budgets/:budget_id/categories/:category_group_id", server.createCategory)
		beta_users.PUT("/budgets/:budget_id/categories/:category_id", server.updateCategory)
		beta_users.DELETE("/budgets/:budget_id/categories/:category_id", server.deleteCategory)

		// payees
		beta_users.GET("/budgets/:budget_id/payees", server.getPayees)
		beta_users.GET("/budgets/:budget_id/payees/:payee_id", server.getPayee)
		beta_users.POST("/budgets/:budget_id/payees", server.createPayee)
		beta_users.PUT("/budgets/:budget_id/payees/:payee_id", server.updatePayee)
		beta_users.DELETE("/budgets/:budget_id/payees/:payee_id", server.deletePayee)

		// transactions
		beta_users.GET("/budgets/:budget_id/transactions", server.getTransactions)
		beta_users.GET("/budgets/:budget_id/transactions/:transaction_id", server.getTransaction)
		beta_users.POST("/budgets/:budget_id/transactions", server.createTransaction)
		// beta_users.PUT("/budgets/:budget_id/transactions/:transaction_id", server.updateTransaction)
	}

	// No auth required
	beta_public := Router.Group("beta")
	{
		// Signup flow
		beta_public.POST("/user", server.createUser)
		beta_public.GET("/verify_email", server.verifyEmail)

		beta_public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

		beta_public.POST("/login", server.login)
		beta_public.POST("/renew_token", server.renewToken)
	}
	server.Router = Router

	return server, nil
}

func errorResponse(msg string) gin.H {
	return gin.H{
		"error": msg,
	}
}
