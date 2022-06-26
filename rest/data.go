package rest

type TempSession struct {
	SessionId     string `json:"sessionId"`
	CodeChallenge string `json:"codeChallenge"`
}

type Session struct {
	SessionId   string `json:"sessionId"`
	ExpiredTime string `json:"expiredTime"`
	IsNew       bool   `json:"isNew"`
}

type TwitterToken struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

type TwitterUserData struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	UserName string `json:"username"`
}

type TwitterUser struct {
	Data TwitterUserData `json:"data"`
}
