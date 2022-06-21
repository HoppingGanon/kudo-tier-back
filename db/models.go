package db

import (
	"time"
)

type TempSession struct {
	SessionID    string    `gorm:"primaryKey;not null"`
	AccessTime   time.Time `gorm:"not null"`
	IpAddress    string    `gorm:"not null;default:0.0.0.0"`
	CodeVerifier string    `gorm:"not null"`
}

type Session struct {
	SessionID    string    `gorm:"primaryKey;not null"`
	AccessTime   time.Time `gorm:"not null"`
	IpAddress    string    `gorm:"not null;default:0.0.0.0"`
	TwitterToken string    `gorm:"not null"`
}

type User struct {
	Id             string    `gorm:"primaryKey;not null"`
	TwitterId      string    `gorm:"index:unique;not null"`
	CreationTime   time.Time `gorm:"not null"`
	LastAccessTime time.Time `gorm:"not null"`
	Name           string    `gorm:"not null;default:no name"`
	IsNew          string    `gorm:"not null;default:true"`
}

type AccessLog struct {
	AccessTime time.Time `gorm:"not null"`
	UserId     string    `gorm:"not null"`
	IpAddress  string    `gorm:"not null;default:0.0.0.0"`
	Operation  string    `gorm:"not null"`
}
