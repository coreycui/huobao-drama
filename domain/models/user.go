package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"username"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Role         string    `gorm:"type:varchar(20);not null;default:'user'" json:"role"`
	IsActive     bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

func (u *User) TableName() string {
	return "users"
}
