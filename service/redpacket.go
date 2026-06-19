package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"redpacket/database"
	"redpacket/models"
	"time"

	"gorm.io/gorm"
)

type CreateActivityReq struct {
	Name        string    `json:"name" binding:"required"`
	CreatorID   string    `json:"creator_id"`
	TotalAmount int64     `json:"total_amount" binding:"required,min=1"`
	TotalCount  int       `json:"total_count" binding:"required,min=1"`
	MinAmount   int64     `json:"min_amount" binding:"required,min=1"`
	MaxAmount   int64     `json:"max_amount" binding:"required,min=1"`
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

	creatorID := req.CreatorID
	if creatorID == "" {
		creatorID = "system"
	}

	activity := &models.RedPacketActivity{
		Name:            req.Name,
		CreatorID:       creatorID,
		TotalAmount:     req.TotalAmount,
		TotalCount:      req.TotalCount,
		RemainingAmount: req.TotalAmount,
		RemainingCount:  req.TotalCount,
		RefundedAmount:  0,
		MinAmount:       req.MinAmount,
		MaxAmount:       req.MaxAmount,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		Status:          models.ActivityStatusActive,
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
		if activity.Status != models.ActivityStatusActive {
			return errors.New("activity is not active")
		}
		if activity.RemainingCount <= 0 {
			return errors.New("red packets have been exhausted")
		}
		if activity.RemainingAmount <= 0 {
			return errors.New("red packet funds have been exhausted")
		}
		if activity.RemainingAmount < activity.MinAmount {
			return errors.New("remaining amount is less than minimum amount, cannot grab")
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

		if amount > activity.RemainingAmount {
			return errors.New("calculated amount exceeds remaining amount")
		}
		if amount < activity.MinAmount && activity.RemainingCount > 1 {
			return errors.New("calculated amount is less than minimum amount")
		}

		result := tx.Model(&models.RedPacketActivity{}).
			Where("id = ? AND remaining_count >= 1 AND remaining_amount >= ?", activity.ID, amount).
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

		var updated models.RedPacketActivity
		if err := tx.First(&updated, activity.ID).Error; err != nil {
			return err
		}
		if updated.RemainingAmount < 0 {
			return errors.New("remaining amount went negative, rolling back")
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
		if remainingAmount > maxAmount {
			return maxAmount, nil
		}
		return remainingAmount, nil
	}

	upperBound := maxAmount
	maxPossible := remainingAmount - int64(remainingCount-1)*minAmount
	if upperBound > maxPossible {
		upperBound = maxPossible
	}

	lowerBound := minAmount
	minPossible := remainingAmount - int64(remainingCount-1)*maxAmount
	if minPossible > minAmount {
		lowerBound = minPossible
	}

	if lowerBound > upperBound {
		upperBound = lowerBound
	}

	if upperBound <= 0 {
		return 0, errors.New("no valid amount can be calculated, remaining pool insufficient")
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

func generateRefundOrderNo() string {
	now := time.Now()
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("RF%s%06d", now.Format("20060102150405"), n.Int64())
}

func FindExpiredActivities() ([]models.RedPacketActivity, error) {
	now := time.Now()
	var activities []models.RedPacketActivity
	err := database.DB.
		Where("end_time < ? AND status = ? AND remaining_amount > 0", now, models.ActivityStatusActive).
		Or("end_time < ? AND status = ? AND remaining_count > 0", now, models.ActivityStatusActive).
		Find(&activities).Error
	return activities, err
}

func ProcessActivityRefund(activityID uint) (*models.RefundRecord, error) {
	var refundRecord *models.RefundRecord

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var activity models.RedPacketActivity
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&activity, activityID).Error; err != nil {
			return errors.New("activity not found")
		}

		if activity.Status == models.ActivityStatusRefunded {
			return errors.New("activity has already been refunded")
		}
		if activity.RemainingAmount <= 0 {
			now := time.Now()
			result := tx.Model(&activity).Updates(map[string]interface{}{
				"status":      models.ActivityStatusClosed,
				"refunded_at": now,
			})
			if result.Error != nil {
				return result.Error
			}
			return errors.New("no remaining amount to refund, activity closed")
		}

		now := time.Now()
		refundAmount := activity.RemainingAmount

		refundRecord = &models.RefundRecord{
			ActivityID:    activity.ID,
			CreatorID:     activity.CreatorID,
			RefundAmount:  refundAmount,
			RefundOrderNo: generateRefundOrderNo(),
			Status:        models.RefundStatusPending,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := tx.Create(refundRecord).Error; err != nil {
			return err
		}

		refundRecord.Status = models.RefundStatusSuccess
		refundRecord.RefundedAt = &now
		refundRecord.UpdatedAt = now
		if err := tx.Save(refundRecord).Error; err != nil {
			return err
		}

		result := tx.Model(&activity).
			Where("id = ? AND status = ?", activity.ID, models.ActivityStatusActive).
			Updates(map[string]interface{}{
				"remaining_amount": int64(0),
				"remaining_count":  int(0),
				"refunded_amount":  refundAmount,
				"status":           models.ActivityStatusRefunded,
				"refunded_at":      now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("failed to update activity status, refund rolled back")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return refundRecord, nil
}

func RunAutoRefund() (int, int64, error) {
	activities, err := FindExpiredActivities()
	if err != nil {
		return 0, 0, err
	}

	successCount := 0
	var totalRefundAmount int64 = 0

	for i := range activities {
		activity := activities[i]
		refundRec, err := ProcessActivityRefund(activity.ID)
		if err != nil {
			continue
		}
		successCount++
		totalRefundAmount += refundRec.RefundAmount
	}

	return successCount, totalRefundAmount, nil
}

func StartAutoRefundScheduler(intervalSeconds int, stopCh <-chan struct{}) {
	go func() {
		log.Println("[RefundScheduler] Started, interval:", intervalSeconds, "seconds")

		go func() {
			time.Sleep(2 * time.Second)
			log.Println("[RefundScheduler] Running initial refund scan...")
			count, amount, err := RunAutoRefund()
			if err != nil {
				log.Printf("[RefundScheduler] Initial scan error: %v", err)
			} else {
				log.Printf("[RefundScheduler] Initial scan done: refunded %d activities, total amount: %d", count, amount)
			}
		}()

		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				log.Println("[RefundScheduler] Stopped")
				return
			case <-ticker.C:
				log.Println("[RefundScheduler] Running scheduled refund scan...")
				count, amount, err := RunAutoRefund()
				if err != nil {
					log.Printf("[RefundScheduler] Scheduled scan error: %v", err)
				} else if count > 0 {
					log.Printf("[RefundScheduler] Scheduled scan done: refunded %d activities, total amount: %d", count, amount)
				} else {
					log.Println("[RefundScheduler] Scheduled scan done: no expired activities found")
				}
			}
		}
	}()
}

func GetRefundRecords(activityID *uint, creatorID string, page, pageSize int) ([]models.RefundRecord, int64, error) {
	var records []models.RefundRecord
	var total int64

	query := database.DB.Model(&models.RefundRecord{})
	if activityID != nil {
		query = query.Where("activity_id = ?", *activityID)
	}
	if creatorID != "" {
		query = query.Where("creator_id = ?", creatorID)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}
