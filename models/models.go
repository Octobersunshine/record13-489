package models

import (
	"time"
)

type RedPacketActivity struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"size:255;not null" json:"name"`
	TotalAmount     int64     `gorm:"not null" json:"total_amount"`
	TotalCount      int       `gorm:"not null" json:"total_count"`
	RemainingAmount int64     `gorm:"not null" json:"remaining_amount"`
	RemainingCount  int       `gorm:"not null" json:"remaining_count"`
	MinAmount       int64     `gorm:"not null" json:"min_amount"`
	MaxAmount       int64     `gorm:"not null" json:"max_amount"`
	StartTime       time.Time `gorm:"not null" json:"start_time"`
	EndTime         time.Time `gorm:"not null" json:"end_time"`
	Status          int       `gorm:"default:1" json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type RedPacketRecord struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ActivityID  uint      `gorm:"index;not null" json:"activity_id"`
	UserID      string    `gorm:"size:64;index;not null" json:"user_id"`
	Amount      int64     `gorm:"not null" json:"amount"`
	OrderNo     string    `gorm:"size:64;uniqueIndex;not null" json:"order_no"`
	CreatedAt   time.Time `json:"created_at"`
	Activity    RedPacketActivity `gorm:"foreignKey:ActivityID" json:"-"`
}

type UserRedPacket struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     string    `gorm:"size:64;index;not null" json:"user_id"`
	ActivityID uint      `gorm:"index;not null" json:"activity_id"`
	RecordID   uint      `gorm:"not null" json:"record_id"`
	Amount     int64     `gorm:"not null" json:"amount"`
	Status     int       `gorm:"default:1" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UsedAt     *time.Time `json:"used_at,omitempty"`
}
