package common

type TwitterToken struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

type TwitterToken1 struct {
	OAuthToken       string `json:"oauth_token"`        // アクセストークン
	OAuthTokenSecret string `json:"oauth_token_secret"` // アクセストークンシークレット
	UserId           string `json:"user_id"`            // @名
	ScreenName       string `json:"screen_name"`        // 表示名
}

type OAuth1RequestToken struct {
	Path          string `json:"path"`
	RequestToken  string `json:"request_token"`
	RequestSecret string `json:"request_secret"`
}

// ユーザーに送付する一時セッションと認証に必要な情報のペア
type TempSessionData struct {
	SessionId string `json:"sessionId"` // OA1, OA2 セッションID
	Url       string `json:"url"`       // OA1, OA2 ユーザーが認証するためのページ
}

type SessionData struct {
	SessionId   string `json:"sessionId"`
	UserId      string `json:"userId"`
	ExpiredTime string `json:"expiredTime"`
	IsNew       bool   `json:"isNew"`
	IconUrl     string `json:"iconUrl"`

	TwitterUserName string `json:"twitterUserName"` // @名
	TwitterName     string `json:"twitterName"`     // 表示名
	TwitterIconUrl  string `json:"twitterIconUrl"`  // アイコン

	GoogleEmail    string `json:"googleEmail"`
	GoogleImageUrl string `json:"googleImageUrl"`
}

type GoogleInfoData struct {
	Id            string `json:"id"`             // ユーザーID
	Email         string `json:"email"`          // メールアドレス
	VerifiedEmail bool   `json:"verified_email"` //
	Name          string `json:"name"`           // フルネーム
	GivenName     string `json:"given_name"`     // 名前
	FamilyName    string `json:"family_name"`    // 苗字
	Picture       string `json:"picture"`        // プロフィール画像
	Local         string `json:"locale"`         // 国
}

// ユーザーから返却される一時セッションと認証情報のペア
type ClientTempSession struct {
	SessionId         string `json:"sessionId"`         // OA1, OA2 セッションID
	AuthorizationCode string `json:"authorizationCode"` // OA2 認証コード
	State             string `json:"state"`             // OA2 発行元をチェックするためのstate
	OAuthToken        string `json:"oAuthToken"`        // OA1 発行済みのトークン
	OAuthVerifier     string `json:"oAuthVerifier"`     // OA1 バックエンドで検証するコード
}
