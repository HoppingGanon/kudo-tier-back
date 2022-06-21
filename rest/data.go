package rest

type TempSession struct {
	SessionId     string `json:"session_id"`
	CodeChallenge string `json:"code_challenge"`
}

type Session struct {
	SessionId   string `json:"session_id"`
	ExpiredTime string `json:"expired_time"`
}

type TwitterToken struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

type TwitterUser struct {
	Data TwitterUserData `json:"data"`
}

type TwitterUserData struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	UserName string `json:"username"`
}
