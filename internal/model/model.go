package model

import "time"

// 短链模型：
// - code 与 long_url 均为唯一索引
// - 记录创建时间
type ShortURL struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	Code      string    `gorm:"size:16;uniqueIndex;not null"`
	LongURL   string    `gorm:"type:varchar(512);uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// 指定表名
func (ShortURL) TableName() string { return "short_urls" }
