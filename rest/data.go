package rest

type ErrorResponse struct {
	/*
		エラーコード
			xxxx-nnn[-ee]
			xxxx 機能
				 gen0 一般的なエラー(セッションなど)
				 ses0 セッションエラー
				 vtir Tierのバリデーションエラー
				 ptir PostTiierのエラー
				 gtir GetTiierのエラー
				 grev GetReviewのエラー
			nnn 項目番号
			ee エラー詳細番号(特に詳細がなければ省略)
	*/
	Code string `json:"code"`
	// エラーメッセージ
	Message string `json:"message"`
}

// ユーザーに送付する一時セッションと認証に必要な情報のペア
type TempSession struct {
	SessionId string `json:"sessionId"` // OA1, OA2 セッションID
	Url       string `json:"url"`       // OA1, OA2 ユーザーが認証するためのページ
}

// ユーザーから返却される一時セッションと認証情報のペア
type ClientTempSession struct {
	SessionId         string `json:"sessionId"`         // OA1, OA2 セッションID
	AuthorizationCode string `json:"authorizationCode"` // OA2 認証コード
	Service           string `json:"service"`           // OA1, OA2 連携サービス
	Version           string `json:"version"`           // OA1, OA2 OAuth認証バージョン
	State             string `json:"state"`             // OA2 発行元をチェックするためのstate
	OAuthToken        string `json:"oAuthToken"`        // OA1 発行済みのトークン
	OAuthVerifier     string `json:"oAuthVerifier"`     // OA1 TierReviewsバックエンドで検証するコード
}

type Session struct {
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

type TwitterUserData struct {
	Id              string `json:"id"`                // 固有ID
	UserName        string `json:"username"`          // @名
	Name            string `json:"name"`              // 表示名
	ProfileImageUrl string `json:"profile_image_url"` // アイコン
}
type TwitterUser struct {
	Data TwitterUserData `json:"data"`
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

type UserCreatingData struct {
	Name       string `json:"name"`       // 登録名
	Profile    string `json:"profile"`    // 自己紹介文
	Accept     bool   `json:"accept"`     // 利用規約への同意(初回のみ)
	IconBase64 string `json:"iconBase64"` // アイコンデータbase64
}

type UserEditingData struct {
	Name             string `json:"name"`             // 登録名
	Profile          string `json:"profile"`          // 自己紹介文
	IconBase64       string `json:"iconBase64"`       // アイコンデータbase64
	IconIsChanged    bool   `json:"iconIsChanged"`    // アイコンが変更されているかどうか
	AllowTwitterLink bool   `json:"allowTwitterLink"` // Twitterへのリンク許可
	KeepSession      int    `json:"keepSession"`      // セッション保持時間(自分自身でのログイン時のみ開示)
}

type UserData struct {
	UserId           string `json:"userId"`           // ユーザーID
	IsSelf           bool   `json:"isSelf"`           // ログインしている自分自身のデータかどうか
	IconUrl          string `json:"iconUrl"`          // アイコンURL
	Name             string `json:"name"`             // 登録名
	Profile          string `json:"profile"`          // 自己紹介文
	AllowTwitterLink bool   `json:"allowTwitterLink"` // Twitterへのリンク許可
	TwitterId        string `json:"twitterId"`        // TwitterID(自分自身でのログイン時およびTwitter連携を許可した時のみ開示)
	ReviewsCount     int64  `json:"reviewsCount"`     // 今までに投稿したレビュー数
	TiersCount       int64  `json:"tiersCount"`       // 今までに投稿したTier数
}

type SelfUserData struct {
	UserId           string `json:"userId"`           // ユーザーID
	IsSelf           bool   `json:"isSelf"`           // ログインしている自分自身のデータかどうか
	IconUrl          string `json:"iconUrl"`          // アイコンURL
	Name             string `json:"name"`             // 登録名
	Profile          string `json:"profile"`          // 自己紹介文
	AllowTwitterLink bool   `json:"allowTwitterLink"` // Twitterへのリンク許可
	KeepSession      int    `json:"keepSession"`      // セッション保持時間(自分自身でのログイン時のみ開示)
	TwitterId        string `json:"twitterId"`        // TwitterID(自分自身でのログイン時およびTwitter連携を許可した時のみ開示)
	TwitterUserName  string `json:"twitterUserName"`  // Twitter@名(自分自身でのログイン時のみ開示)
	GoogleEmail      string `json:"googleEmail"`      // Google Mailアドレス(自分自身でのログイン時のみ開示)
	ReviewsCount     int64  `json:"reviewsCount"`     // 今までに投稿したレビュー数
	TiersCount       int64  `json:"tiersCount"`       // 今までに投稿したTier数
}

type TierData struct {
	TierId             string            `json:"tierId"`
	UserName           string            `json:"userName"`
	UserId             string            `json:"userId"`
	UserIconUrl        string            `json:"userIconUrl"`
	Name               string            `json:"name"`
	ImageUrl           string            `json:"imageUrl"`
	Parags             []ParagData       `json:"parags"`
	Reviews            []ReviewData      `json:"reviews"`
	PointType          string            `json:"pointType"`
	ReviewFactorParams []ReviewParamData `json:"reviewFactorParams"`
	PullingUp          int               `json:"pullingUp"`
	PullingDown        int               `json:"pullingDown"`
	CreatedAt          string            `json:"createdAt"`
	UpdatedAt          string            `json:"updatedAt"`
}

type ReviewData struct {
	ReviewId      string             `json:"reviewId"`
	UserName      string             `json:"userName"`
	UserId        string             `json:"userId"`
	UserIconUrl   string             `json:"userIconUrl"`
	TierId        string             `json:"tierId"`
	Title         string             `json:"title"`
	Name          string             `json:"name"`
	IconUrl       string             `json:"iconUrl"`
	ReviewFactors []ReviewFactorData `json:"reviewFactors"`
	PointType     string             `json:"pointType"`
	Sections      []SectionData      `json:"sections"`
	CreatedAt     string             `json:"createdAt"`
	UpdatedAt     string             `json:"updatedAt"`
}

type ReviewDataWithParams struct {
	Review      ReviewData        `json:"review"`
	Params      []ReviewParamData `json:"params"`
	PullingDown int               `json:"pullingDown"`
	PullingUp   int               `json:"pullingUp"`
}

type TierEditingData struct {
	Name               string             `json:"name"`
	ImageBase64        string             `json:"imageBase64"`
	ImageIsChanged     bool               `json:"imageIsChanged"`
	Parags             []ParagEditingData `json:"parags"`
	PointType          string             `json:"pointType"`
	ReviewFactorParams []ReviewParamData  `json:"reviewFactorParams"`
	PullingUp          int                `json:"pullingUp"`
	PullingDown        int                `json:"pullingDown"`
}

type ReviewEditingData struct {
	TierId        string               `json:"tierId"`
	Title         string               `json:"title"`
	Name          string               `json:"name"`
	IconBase64    string               `json:"iconBase64"`
	IconIsChanged bool                 `json:"iconIsChanged"`
	ReviewFactors []ReviewFactorData   `json:"reviewFactors"`
	Sections      []SectionEditingData `json:"sections"`
}

type ReviewFactorData struct {
	Info  string  `json:"info"`
	Point float64 `json:"point"`
}

type ReviewParamData struct {
	Name    string `json:"name"`
	IsPoint bool   `json:"isPoint"`
	Weight  int    `json:"weight"`
	Index   int    `json:"index"`
}

type ReviewParam struct {
	Name    string `json:"name"`
	IsPoint bool   `json:"isPoint"`
	Weight  int    `json:"weight"`
}

type SectionData struct {
	Title  string      `json:"title"`
	Parags []ParagData `json:"parags"`
}

type SectionEditingData struct {
	Title  string             `json:"title"`
	Parags []ParagEditingData `json:"parags"`
}

type ParagEditingData struct {
	Type      string `json:"type"`
	Body      string `json:"body"`
	IsChanged bool   `json:"isChanged"`
}

type ParagData struct {
	Type string `json:"type"`
	Body string `json:"body"`
}

type PostListsData struct {
	Tiers   []PostListItem `json:"tiers"`
	Reviews []PostListItem `json:"reviews"`
}

type PostListItem struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Notification struct {
	Id          uint   `json:"id"`
	Content     string `json:"content"`
	IsRead      bool   `json:"isRead"`
	IsImportant bool   `json:"isImportant"`
	FromUserId  string `json:"fromUserId"`
	Url         string `json:"url"`
	CreatedAt   string `json:"createdAt"`
}

type CountData struct {
	Count int64 `json:"count"`
}

type NotificationReadData struct {
	IsRead bool `json:"isRead"`
}
