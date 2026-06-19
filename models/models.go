package models

import (
	"time"
)

const (
	ActivityStatusActive   = 1
	ActivityStatusRefunded = 2
	ActivityStatusClosed   = 3
)

const (
	RefundStatusPending  = 0
	RefundStatusSuccess  = 1
	RefundStatusFailed   = 2
)

type RedPacketActivity struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"size:255;not null" json:"name"`
	CreatorID       string    `gorm:"size:64;index;not null;default:'system'" json:"creator_id"`
	TotalAmount     int64     `gorm:"not null" json:"total_amount"`
	TotalCount      int       `gorm:"not null" json:"total_count"`
	RemainingAmount int64     `gorm:"not null" json:"remaining_amount"`
	RemainingCount  int       `gorm:"not null" json:"remaining_count"`
	RefundedAmount  int64     `gorm:"not null;default:0" json:"refunded_amount"`
	MinAmount       int64     `gorm:"not null" json:"min_amount"`
	MaxAmount       int64     `gorm:"not null" json:"max_amount"`
	StartTime       time.Time `gorm:"not null" json:"start_time"`
	EndTime         time.Time `gorm:"not null" json:"end_time"`
	Status          int       `gorm:"default:1;index" json:"status"`
	RefundedAt      *time.Time `json:"refunded_at,omitempty"`
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

type RefundRecord struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ActivityID     uint      `gorm:"index;not null" json:"activity_id"`
	CreatorID      string    `gorm:"size:64;index;not null" json:"creator_id"`
	RefundAmount   int64     `gorm:"not null" json:"refund_amount"`
	RefundOrderNo  string    `gorm:"size:64;uniqueIndex;not null" json:"refund_order_no"`
	Status         int       `gorm:"default:0;index" json:"status"`
	FailureReason  string    `gorm:"size:500" json:"failure_reason,omitempty"`
	RefundedAt     *time.Time `json:"refunded_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Activity       RedPacketActivity `gorm:"foreignKey:ActivityID" json:"-"`
}
