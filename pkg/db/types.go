package db

// Parameter with just a username and email
type UserParams struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type UserTxResult struct {
	User User
}
