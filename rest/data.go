package rest

type TempSession struct {
	SessionId     string `json:"session_id"`
	CodeChallenge string `json:"code_challenge"`
}

type TwitterToken struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in"`
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}
