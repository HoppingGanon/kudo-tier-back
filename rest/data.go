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

type TwitterUserData struct {
	Id              string `json:"id"`                // 固有ID
	UserName        string `json:"username"`          // @名
	Name            string `json:"name"`              // 表示名
	ProfileImageUrl string `json:"profile_image_url"` // アイコン
}

type TwitterUser struct {
	Data TwitterUserData `json:"data"`
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
