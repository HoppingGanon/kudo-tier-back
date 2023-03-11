package db

import (
	"time"
)

// 一時セッション
// Twitter OAuth1.0a, 2.0認証が完了するまでの間、フロントとバックの間でデータを共有するために用いる
type TempSession struct {
	SessionID  string    `gorm:"primaryKey;not null"`      // 一時セッションID
	AccessTime time.Time `gorm:"not null"`                 // アクセスした時間
	IpAddress  string    `gorm:"not null;default:0.0.0.0"` // セッション確立時のIPアドレス

	LoginService  string `gorm:"not null"` // ログインに使用したサービス
	LoginVersion  int    `gorm:"not null"` // ログインに使用したOAuthバージョン
	CodeVerifier  string `gorm:""`         // OA2 TwitterのOAuth2.0認証でコード検証に用いるハッシュの生成元文字列
	State         string `gorm:""`         // OA2 OAuth2.0認証でコード検証に用いるstate
	RequestToken  string `gorm:""`         // OA1 OAuth1.0a認証で認証サーバーから受け取るトークン
	RequestSecret string `gorm:""`         // OA1 OAuth1.0a認証で認証サーバーから受け取るハッシュ
}

// セッション
type Session struct {
	SessionId   string    `gorm:"primaryKey;not null"` // セッションID
	UserId      string    `gorm:""`                    // ユーザーデータのID
	ExpiredTime time.Time `gorm:"not null"`            //セッションの有効期限

	LoginService string `gorm:"not null"` // ログインに使用したサービス
	LoginVersion int    `gorm:"not null"` // ログインに使用したサービス

	ServiceId string `gorm:"not null"` // 連携サービスンの固有ID(OA1 OA2)

	TwitterIconUrl  string `gorm:""` // Twitter アイコンURL
	TwitterUserName string `gorm:""` // Twitter @名
	TwitterToken    string `gorm:""` // Twitterから与えられたアクセストークン(OA2)
	TwitterToken1   string `gorm:""` // Twitterから与えられたアクセストークン(OA1)
	TwitterSecret1  string `gorm:""` // Twitterから与えられたアクセスシークレット(OA1)

	GoogleEmail        string    `gorm:""` // Google Email
	GoogleImageUrl     string    `gorm:""` // Google 画像
	GoogleAccessToken  string    `gorm:""` // Googleから与えられたアクセストークン
	GoogleExpiry       time.Time `gorm:""` // Googleから与えられたアクセストークンの期限
	GoogleRefreshToken string    `gorm:""` // Googleから与えられたアクセストークンのリフレッシュ用

	IsNew          bool      `gorm:"not null"`  // ユーザー未登録状態フラグ
	LastPostAt     time.Time `gorm:"not null;"` // 直近の投稿時間
	DeleteCodeTime time.Time `gorm:""`          // ユーザーを削除する際の確認コード生成時間
}

// ユーザーデータ
type User struct {
	UserId           string `gorm:"primaryKey;not null"`    // ランダムで決定するユーザー固有のID
	IconUrl          string `gorm:"not null"`               // TwitterのアイコンURL
	Name             string `gorm:"not null"`               // 登録名
	Profile          string `gorm:"not null"`               // 自己紹介文
	AllowTwitterLink bool   `gorm:"not null;default:false"` // Twitterへのリンク許可
	KeepSession      int    `gorm:"not null;default:7200"`  // セッション保持時間(秒)

	TwitterId       string `gorm:""` // TwitterID(自分自身でのログイン時およびTwitter連携を許可した時のみ開示)
	TwitterUserName string `gorm:""` // @名
	GoogleId        string `gorm:""` // Google 固有ID
	GoogleEmail     string `gorm:""` // Google Gmailアドレス

	CreatedAt time.Time `gorm:""` // 作成日
	UpdatedAt time.Time `gorm:""` // 更新日
}

// アクセスログ
// 条件: ログイン、ログアウト、ユーザー登録・変更・削除、Tier作成・編集・削除、レビュー作成・編集・削除
type OperationLog struct {
	UserId    string    `gorm:"not null"`                 // ユーザーデータの固有ID
	IpAddress string    `gorm:"not null;default:0.0.0.0"` // セッション確立時のIPアドレス
	Operation string    `gorm:"not null"`                 // 操作対象(エラーコードに準じる)
	Content   string    `gorm:"not null"`                 // 操作内容
	CreatedAt time.Time `gorm:"not null;index"`           // 作成日
}

// エラーログ
// 条件: 致命的なエラーの場合
type ErrorLog struct {
	UserId       string    `gorm:"not null"`                 // ユーザーデータの固有ID
	IpAddress    string    `gorm:"not null;default:0.0.0.0"` // セッション確立時のIPアドレス
	ErrorId      string    `gorm:"not null"`                 // エラーID
	Operation    string    `gorm:"not null"`                 // 操作内容
	Descriptions string    `gorm:"not null"`                 // 操作内容(詳細)
	CreatedAt    time.Time `gorm:"not null;index"`           // 作成日
}

// Tier
type Tier struct {
	TierId       string    `gorm:"primaryKey;not null"` // Tier固有のID
	UserId       string    `gorm:"not null;index"`      // 作成ユーザーの固有ID
	Name         string    `gorm:"not null"`            // Tierの名称
	ImageUrl     string    `gorm:"not null"`            // Tierカバー画像のURL
	Parags       string    `gorm:"not null"`            // 説明文
	PointType    string    `gorm:"not null"`            // デフォルトのポイント表示形式
	FactorParams string    `gorm:"not null"`            // 評価のパラメータ
	PullingUp    int       `gorm:"not null"`            // Tier表を上に引き上げる
	PullingDown  int       `gorm:"not null"`            // Tier表を下に引き下げる
	CreatedAt    time.Time `gorm:""`                    // 作成日
	UpdatedAt    time.Time `gorm:""`                    // 更新日
}

// Review
type Review struct {
	ReviewId      string    `gorm:"primaryKey;not null"` // レビュー固有のID
	UserId        string    `gorm:"not null"`            // 作成ユーザーの固有ID
	TierId        string    `gorm:"not null"`            // 作成元Tierの固有ID
	Title         string    `gorm:"not null"`            // レビューのタイトル
	Name          string    `gorm:"not null"`            // レビューの名前
	IconUrl       string    `gorm:"not null"`            // レビューアイコンのURL
	ReviewFactors string    `gorm:"not null"`            // レビューの評価要素
	Sections      string    `gorm:"not null"`            // レビュー説明セクション
	CreatedAt     time.Time `gorm:""`                    // 作成日
	UpdatedAt     time.Time `gorm:""`                    // 更新日
}

type Notification struct {
	Id          uint      `gorm:"primaryKey"`
	Content     string    `gorm:""`                       // 表示する文章
	IsImportant bool      `gorm:"default:false;not null"` // 重要情報フラグ
	Url         string    `gorm:""`                       // クリックした際に飛ぶURL
	CreatedAt   time.Time `gorm:"index"`                  // 発信日時
}

type NotificationRead struct {
	NotificationId uint   `gorm:"primaryKey"` // NotificationのID
	UserId         string `gorm:"primaryKey"` // ユーザーID
	IsRead         bool   `gorm:"not null"`   // 既読フラグ
}

type NotificationJoinRead struct {
	Id          uint
	Content     string
	IsRead      bool
	IsImportant bool
	Url         string
	CreatedAt   time.Time
}
