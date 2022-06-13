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
	SessionID  string    `gorm:"primaryKey;not null"`
	AccessTime time.Time `gorm:"not null"`
	IpAddress  string    `gorm:"not null;default:0.0.0.0"`
	TwitterId  string    `gorm:"not null;default:none"`
}

type User struct {
	TwitterId  string    `gorm:"primaryKey;not null"`
	Jwt        string    `gorm:"index:idx_name,unique"`
	AccessTime time.Time `gorm:"not null"`
	IpAddress  string    `gorm:"not null;default:0.0.0.0"`
}
