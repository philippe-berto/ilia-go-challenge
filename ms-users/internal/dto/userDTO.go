package dto

type UserOutput struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type AuthOutput struct {
	User        *UserOutput `json:"user"`
	AccessToken string      `json:"access_token"`
}
