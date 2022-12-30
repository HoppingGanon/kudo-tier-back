package rest

type ErrorResponse struct {
	/*
		エラーコード
			xxxx-nn[-ee]
			xxxx 機能
				gen0 一般的なエラー(セッションなど)
				ses0 セッションエラー
				vtir Tierのバリデーションエラー
				ptir PostTiierのエラー
				gtir GetTiierのエラー
				grev GetReviewのエラー
			nn 項目番号
			ee エラー詳細番号(特に詳細がなければ省略)
	*/
	Code string `json:"code"`
	// エラーメッセージ
	Message string `json:"message"`
}

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
	CreateAt           string            `json:"createAt"`
	UpdatedAt          string            `json:"updateAt"`
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
	CreateAt      string             `json:"createAt"`
	UpdateAt      string             `json:"updateAt"`
}

type TierEditingData struct {
	TierId             string            `json:"tierId"`
	Name               string            `json:"name"`
	ImageBase64        string            `json:"imageBase64"`
	Parags             []ParagData       `json:"parags"`
	PointType          string            `json:"pointType"`
	ReviewFactorParams []ReviewParamData `json:"reviewFactorParams"`
}

type ReviewFactorData struct {
	Info  string `json:"info"`
	Point int    `json:"point"`
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

type ParagData struct {
	Type string `json:"type"`
	Body string `json:"body"`
}
