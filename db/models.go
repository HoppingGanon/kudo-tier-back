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
}

// ユーザーデータ
type User struct {
	UserId      string    `gorm:"primaryKey;not null"`       // ランダムで決定するユーザー固有のID
	TwitterName string    `gorm:"index:unique;not null"`     // ユーザーのTwitterID
	IconUrl     string    `gorm:"not null;default:no image"` // TwitterのアイコンURL
	Name        string    `gorm:"not null;default:no name"`  // 登録名
	Profile     string    `gorm:"not null"`                  // 自己紹介文
	CreatedAt   time.Time `gorm:""`                          // 作成日
	UpdatedAt   time.Time `gorm:""`                          // 更新日
	DeletedAt   time.Time `gorm:"index"`                     // 削除日
}

// アクセスログ
// 条件: ログイン、ログアウト、ユーザー登録・変更・削除、記事作成・編集・削除、コメント追加・編集・削除
type OperationLog struct {
	UserId     string    `gorm:"not null"`                 // ユーザーデータの固有ID
	IpAddress  string    `gorm:"not null;default:0.0.0.0"` // セッション確立時のIPアドレス
	Operation  string    `gorm:"not null"`                 // 操作内容
	AccessTime time.Time `gorm:"not null"`                 // ログを記録した時間
}
