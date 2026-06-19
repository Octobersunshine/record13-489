package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"redpacket/database"
	"redpacket/models"
	"time"

	"gorm.io/gorm"
)

type CreateActivityReq struct {
	Name      string    `json:"name" binding:"required"`
	TotalAmount int64   `json:"total_amount" binding:"required,min=1"`
	TotalCount  int     `json:"total_count" binding:"required,min=1"`
	MinAmount   int64   `json:"min_amount" binding:"required,min=1"`
	MaxAmount   int64   `json:"max_amount" binding:"required,min=1"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
}

type GrabRedPacketReq struct {
	ActivityID uint   `json:"activity_id" binding:"required"`
	UserID     string `json:"user_id" binding:"required"`
}

func CreateActivity(req *CreateActivityReq) (*models.RedPacketActivity, error) {
	if req.EndTime.Before(req.StartTime) {
		return nil, errors.New("end_time must be after start_time")
	}
	if req.MinAmount > req.MaxAmount {
		return nil, errors.New("min_amount must be less than or equal to max_amount")
	}
	if int64(req.TotalCount)*req.MinAmount > req.TotalAmount {
		return nil, errors.New("total_amount is insufficient for min_amount per packet")
	}
	if int64(req.TotalCount)*req.MaxAmount < req.TotalAmount {
		return nil, errors.New("total_amount exceeds max_amount per packet capacity")
	}

	activity := &models.RedPacketActivity{
		Name:            req.Name,
		TotalAmount:     req.TotalAmount,
		TotalCount:      req.TotalCount,
		RemainingAmount: req.TotalAmount,
		RemainingCount:  req.TotalCount,
		MinAmount:       req.MinAmount,
		MaxAmount:       req.MaxAmount,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		Status:          1,
	}

	if err := database.DB.Create(activity).Error; err != nil {
		return nil, err
	}
	return activity, nil
}

func GrabRedPacket(req *GrabRedPacketReq) (*models.RedPacketRecord, error) {
	var record *models.RedPacketRecord

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var activity models.RedPacketActivity
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&activity, req.ActivityID).Error; err != nil {
			return errors.New("activity not found")
		}

		now := time.Now()
		if now.Before(activity.StartTime) {
			return errors.New("activity has not started yet")
		}
		if now.After(activity.EndTime) {
			return errors.New("activity has already ended")
		}
		if activity.Status != 1 {
			return errors.New("activity is not active")
		}
		if activity.RemainingCount <= 0 {
			return errors.New("red packets have been exhausted")
		}
		if activity.RemainingAmount <= 0 {
			return errors.New("red packet funds have been exhausted")
		}

		var existingCount int64
		tx.Model(&models.RedPacketRecord{}).
			Where("activity_id = ? AND user_id = ?", req.ActivityID, req.UserID).
			Count(&existingCount)
		if existingCount > 0 {
			return errors.New("you have already claimed this red packet")
		}

		amount, err := calculateAmount(&activity)
		if err != nil {
			return err
		}

		orderNo := generateOrderNo()

		record = &models.RedPacketRecord{
			ActivityID: req.ActivityID,
			UserID:     req.UserID,
			Amount:     amount,
			OrderNo:    orderNo,
			CreatedAt:  now,
		}
		if err := tx.Create(record).Error; err != nil {
			return err
		}

		userRedPacket := &models.UserRedPacket{
			UserID:     req.UserID,
			ActivityID: req.ActivityID,
			RecordID:   record.ID,
			Amount:     amount,
			Status:     1,
			CreatedAt:  now,
		}
		if err := tx.Create(userRedPacket).Error; err != nil {
			return err
		}

		result := tx.Model(&activity).
			Where("remaining_count >= 1 AND remaining_amount >= ?", amount).
			Updates(map[string]interface{}{
				"remaining_count":  gorm.Expr("remaining_count - 1"),
				"remaining_amount": gorm.Expr("remaining_amount - ?", amount),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("failed to grab red packet, please try again")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return record, nil
}

func calculateAmount(activity *models.RedPacketActivity) (int64, error) {
	remainingCount := activity.RemainingCount
	remainingAmount := activity.RemainingAmount
	minAmount := activity.MinAmount
	maxAmount := activity.MaxAmount

	if remainingCount == 1 {
		return remainingAmount, nil
	}

	lowerBound := minAmount
	upperBound := maxAmount

	maxAvg := (remainingAmount - int64(remainingCount-1)*minAmount)
	if upperBound > maxAvg {
		upperBound = maxAvg
	}

	minAvg := (remainingAmount - int64(remainingCount-1)*maxAmount)
	if minAvg < minAmount {
		minAvg = minAmount
	}
	if lowerBound < minAvg {
		lowerBound = minAvg
	}

	if lowerBound > upperBound {
		lowerBound = upperBound
	}

	rangeSize := upperBound - lowerBound + 1
	if rangeSize <= 0 {
		return lowerBound, nil
	}

	n, err := rand.Int(rand.Reader, big.NewInt(rangeSize))
	if err != nil {
		return 0, errors.New("failed to generate random amount")
	}

	amount := lowerBound + n.Int64()

	if amount < minAmount {
		amount = minAmount
	}
	if amount > maxAmount {
		amount = maxAmount
	}
	if amount > remainingAmount {
		amount = remainingAmount
	}

	return amount, nil
}

func generateOrderNo() string {
	now := time.Now()
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("RP%s%06d", now.Format("20060102150405"), n.Int64())
}

func GetActivity(id uint) (*models.RedPacketActivity, error) {
	var activity models.RedPacketActivity
	if err := database.DB.First(&activity, id).Error; err != nil {
		return nil, errors.New("activity not found")
	}
	return &activity, nil
}

func ListActivities(page, pageSize int) ([]models.RedPacketActivity, int64, error) {
	var activities []models.RedPacketActivity
	var total int64

	query := database.DB.Model(&models.RedPacketActivity{})
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&activities).Error; err != nil {
		return nil, 0, err
	}
	return activities, total, nil
}

func GetRecords(activityID uint, page, pageSize int) ([]models.RedPacketRecord, int64, error) {
	var records []models.RedPacketRecord
	var total int64

	query := database.DB.Model(&models.RedPacketRecord{}).Where("activity_id = ?", activityID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func GetUserRedPackets(userID string, page, pageSize int) ([]models.UserRedPacket, int64, error) {
	var packets []models.UserRedPacket
	var total int64

	query := database.DB.Model(&models.UserRedPacket{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&packets).Error; err != nil {
		return nil, 0, err
	}
	return packets, total, nil
}
