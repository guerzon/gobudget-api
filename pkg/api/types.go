package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/jackc/pgx/v5/pgtype"
)

const internal_error_message = "Internal server error. Please contact the administrator."

type createUserRequest struct {
	Username string `json:"username" binding:"required,min=5"`
	Email    string `json:"email" binding:"required,email" example:"fname.lname@contoso.com"`
	Password string `json:"password" binding:"required,min=10" example:"password123456"`
} //@name CreateUserRequest

type updateUserRequest struct {
	Email    pgtype.Text `json:"email" example:"fname.lname@contoso.com" swaggertype:"string"`
	Password pgtype.Text `json:"password" example:"password123456" swaggertype:"string"`
} //@name UpdateUserRequest

// Custom user struct used for user responses
type userResponse struct {
	Username           string    `json:"username" example:"rjoooidggt"`
	Email              string    `json:"email" example:"fname.lname@contoso.com"`
	EmailVerified      bool      `json:"email_verified" example:"true"`
	CreatedAt          time.Time `json:"created_at,omitempty" example:"2023-09-29T22:14:50+08:00"`
	LastPasswordChange time.Time `json:"last_password_change,omitempty" example:"2023-09-29T22:14:50+08:00"`
} //@name UserResponse

// Only used for Swagger
type HTTPError struct {
	Message string `json:"msg" example:"invalid request"`
}

type BudgetId struct {
	BudgetId string `uri:"budget_id" binding:"required,uuid"`
}

type AccountId struct {
	AccountId string `uri:"account_id" binding:"required"`
}

type accountRequest struct {
	Name    string `json:"name" binding:"required" example:"Chase Savings"`
	Type    string `json:"type" binding:"required" example:"Savings"`
	Balance int32  `json:"balance"`
}

type updateAccountRequest struct {
	Name             pgtype.Text        `json:"name" example:"Chase Savings" swaggertype:"string"`
	Type             pgtype.Text        `json:"type" example:"Savings" swaggertype:"string"`
	Closed           pgtype.Bool        `json:"closed" example:"false" swaggertype:"boolean"`
	Note             pgtype.Text        `json:"note" swaggertype:"string"`
	Balance          pgtype.Int4        `json:"balance" example:"100" swaggertype:"integer"`
	ClearedBalance   pgtype.Int4        `json:"cleared_balance" example:"50" swaggertype:"integer"`
	UnclearedBalance pgtype.Int4        `json:"uncleared_balance" example:"50" swaggertype:"integer"`
	LastReconciledAt pgtype.Timestamptz `json:"last_reconciled_at" swaggertype:"string"`
}

type budgetRequest struct {
	Name         string `json:"name" example:"My USD Budget"`
	CurrencyCode string `json:"currency_code" binding:"iso4217" example:"USD"`
}

type detailedBudgetResponse struct {
	Id           uuid.UUID    `json:"id" example:"ea930f68-e192-407d..."`
	Name         string       `json:"name" example:"My USD Budget"`
	CurrencyCode string       `json:"currency_code" example:"USD"`
	Accounts     []db.Account `json:"accounts"`
} //@name DetailedBudgetResponse

type loginRequest struct {
	Username string `json:"username" binding:"required,min=5"`
	Password string `json:"password" binding:"required,min=10"`
}

type loginResponse struct {
	SessionID             uuid.UUID `json:"session_id" example:"ea930f68-e192-407d..."`
	AccessToken           string    `json:"access_token" example:"eyJhbGciOiJIUzI1Ni..."`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at" example:"2023-10-30T22:14:50+08:00"`
	RefreshToken          string    `json:"refresh_token" example:"eyJhbGciOiJIUzI1Ni..."`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at" example:"2023-10-31T22:14:50+08:00"`
	ID                    uuid.UUID `json:"id" example:"ea930f68-e192-407d..."`
	Username              string    `json:"username" example:"rjoooidggt"`
	Email                 string    `json:"email" example:"fname.lname@contoso.com"`
	CreatedAt             time.Time `json:"created_at,omitempty" example:"2023-09-29T22:14:50+08:00"`
	LastPasswordChange    time.Time `json:"last_password_change,omitempty" example:"2023-09-29T22:14:50+08:00"`
} //@name LoginResponse

type renewTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1Ni..."`
} //@name RenewTokenRequest

type renewTokenResponse struct {
	SessionID            uuid.UUID `json:"session_id" example:"ea930f68-e192-407d..."`
	AccessToken          string    `json:"access_token" example:"eyJhbGciOiJIUzI1Ni..."`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at" example:"2023-10-31T22:14:50+08:00"`
} //@name RenewTokenResponse

type verifyEmailRequest struct {
	ID   int64  `form:"id" binding:"required"`
	Code string `form:"code" binding:"required"`
}

type categoryGroupRqst struct {
	Name string `json:"name" binding:"required,min=5" example:"Living Expenses"`
}

type CategoryGroupId struct {
	CategoryGroupId string `uri:"category_group_id" binding:"required,uuid"`
}

type CategoryId struct {
	CategoryId string `uri:"category_id" binding:"required,uuid"`
}

type categoryRqst struct {
	Name string `json:"name" binding:"required,min=2" example:"Rent"`
}

type categoryResponse struct {
	CategoryGroupId uuid.UUID `json:"category_group_id"`
	Name            string    `json:"name" example:"Rent"`
	Categories      []db.Category
}

type PayeeId struct {
	PayeeId string `uri:"payee_id" binding:"required,uuid"`
}

type payeeRqst struct {
	Name string `json:"name" binding:"required,min=2" example:"Edeka"`
}

// Generic UUID type
type TransactionId struct {
	Id string `uri:"transaction_id" binding:"required,uuid"`
}

type transactionRequest struct {
	Account    string      `json:"account_id" binding:"required,uuid" swaggertype:"string"`
	Date       pgtype.Date `json:"date" binding:"required" swaggertype:"string"`
	Payee      string      `json:"payee_id" binding:"required,uuid" swaggertype:"string"`
	Category   string      `json:"category_id" binding:"required,uuid" swaggertype:"string"`
	Memo       string      `json:"memo" swaggertype:"string"`
	Amount     int32       `json:"amount" binding:"required,number"`
	Cleared    bool        `json:"cleared" binding:"boolean"`
	Reconciled bool        `json:"reconciled" binding:"boolean"`
} //@name TransactionRequest

type transactionResponse struct {
	Date       pgtype.Date `json:"date"`
	Account    string      `json:"account_name"`
	Payee      string      `json:"payee_name"`
	Category   string      `json:"category_name"`
	Memo       pgtype.Text `json:"memo"`
	Amount     int32       `json:"amount"`
	Approved   bool        `json:"approved"`
	Cleared    bool        `json:"cleared"`
	Reconciled bool        `json:"reconciled"`
} //@name TransactionResponse
