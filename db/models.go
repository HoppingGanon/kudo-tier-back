package db

import (
	"time"
)

// 一時セッション
// Twitter認証が完了するまでの間、フロントとバックの間でデータを共有するために用いる
type TempSession struct {
	SessionID    string    `gorm:"primaryKey;not null"`      // 一時セッションID
	AccessTime   time.Time `gorm:"not null"`                 // アクセスした時間
	IpAddress    string    `gorm:"not null;default:0.0.0.0"` // セッション確立時のIPアドレス
	CodeVerifier string    `gorm:"not null"`                 // TwitterのOAuth2.0認証でコード検証に用いるハッシュの生成元文字列
}

// セッション
type Session struct {
	SessionID    string    `gorm:"primaryKey;not null"` // セッションID
	UserId       string    `gorm:""`                    // ユーザーデータのID
	ExpiredTime  time.Time `gorm:"not null"`            //セッションの有効期限
	TwitterToken string    `gorm:"not null"`            // Twitterから与えられたトークン
	IsNew        bool      `gorm:"not null"`            // ユーザー未登録状態フラグ
	LastPostAt   time.Time `gorm:"not null;"`           // 直近の投稿時間
}

// ユーザーデータ
type User struct {
	UserId           string    `gorm:"primaryKey;not null"`       // ランダムで決定するユーザー固有のID
	TwitterName      string    `gorm:"index:unique;not null"`     // TwitterID(自分自身でのログイン時およびTwitter連携を許可した時のみ開示)
	IconUrl          string    `gorm:"not null;default:no image"` // TwitterのアイコンURL
	Name             string    `gorm:"not null;default:no name"`  // 登録名
	Profile          string    `gorm:"not null"`                  // 自己紹介文
	AllowTwitterLink bool      `gorm:"not null;default:false"`    // Twitterへのリンク許可
	KeepSession      int       `gorm:"not null;default:3600"`     // セッション保持時間(秒)
	CreatedAt        time.Time `gorm:""`                          // 作成日
	UpdatedAt        time.Time `gorm:""`                          // 更新日
}

// アクセスログ
// 条件: ログイン、ログアウト、ユーザー登録・変更・削除、Tier作成・編集・削除、レビュー作成・編集・削除
type OperationLog struct {
	UserId    string    `gorm:"not null"`                 // ユーザーデータの固有ID
	IpAddress string    `gorm:"not null;default:0.0.0.0"` // セッション確立時のIPアドレス
	Operation string    `gorm:"not null"`                 // 操作対象(エラーコードに準じる)
	Content   string    `gorm:"not null"`                 // 操作内容
	CreatedAt time.Time `gorm:"not null"`                 // 作成日
}

// エラーログ
// 条件: 致命的なエラーの場合
type ErrorLog struct {
	UserId       string    `gorm:"not null"`                 // ユーザーデータの固有ID
	IpAddress    string    `gorm:"not null;default:0.0.0.0"` // セッション確立時のIPアドレス
	ErrorId      string    `gorm:"not null"`                 // エラーID
	Operation    string    `gorm:"not null"`                 // 操作内容
	Descriptions string    `gorm:"not null"`                 // 操作内容(詳細)
	CreatedAt    time.Time `gorm:"not null"`                 // 作成日
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
