package rest

type TempSession struct {
	SessionId     string `json:"sessionId"`
	CodeChallenge string `json:"codeChallenge"`
}

type Session struct {
	SessionId       string `json:"sessionId"`
	UserId          string `json:"userId"`
	ExpiredTime     string `json:"expiredTime"`
	IsNew           bool   `json:"isNew"`
	TwitterName     string `json:"twitterName"`
	TwitterUserName string `json:"twitterUserName"`
	IconUrl         string `json:"iconUrl"`
}

type TwitterToken struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

type TwitterUserData struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	UserName        string `json:"username"`
	ProfileImageUrl string `json:"profile_image_url"`
}
type TwitterUser struct {
	Data TwitterUserData `json:"data"`
}

type InitUserData struct {
	Name    string `json:"name"`    // 登録名
	Profile string `json:"profile"` // 自己紹介文
	Accept  bool   `json:"accept"`  // 利用規約への同意
}

type NewUserData struct {
	Name    string `json:"name"`    // 登録名
	Profile string `json:"profile"` // 自己紹介文
	UserId  string `json:"userId"`  // ユーザーID
	IconUrl string `json:"iconUrl"` // TwitterのアイコンURL
}

type UserData struct {
	IsSelf      bool   `json:"isSelf"`      // ログインしている自分自身のデータかどうか
	IconUrl     string `json:"iconUrl"`     // TwitterのアイコンURL
	Name        string `json:"name"`        // 登録名
	Profile     string `json:"profile"`     // 自己紹介文
	TwitterName string `json:"twitterName"` // TwitterID(ログイン時のみ)
}
