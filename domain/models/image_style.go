package models

import "time"

type ImageStyle struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	NameZH     string    `gorm:"type:varchar(100);not null" json:"name_zh"`
	NameEN     string    `gorm:"type:varchar(100);not null" json:"name_en"`
	StyleValue string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"style_value"`
	SortOrder  int       `gorm:"default:0" json:"sort_order"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

func (ImageStyle) TableName() string {
	return "image_styles"
}
