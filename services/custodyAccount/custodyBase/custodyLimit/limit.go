package custodyLimit

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"sync"
	"time"
	"trade/models/custodyModels"
	caccount "trade/services/custodyAccount/account"
)

const (
	defaultLimitLevel = 1
)

var (
	ErrLimitEntirely = errors.New("ErrLimitEntirely")
)
var limitMux = new(sync.Mutex)

func GetLimit(db *gorm.DB, user *caccount.UserInfo, limitType *custodyModels.LimitType) (*custodyModels.LimitBill, error) {
	err := db.Where(limitType).First(limitType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			return nil, nil
		}

		return nil, err
	}
	limitMux.Lock()
	defer limitMux.Unlock()

	limitBill := custodyModels.LimitBill{
		UserId:    user.User.ID,
		LimitType: limitType.ID,
	}
	err = db.Where("created_at >= CURDATE() AND created_at < CURDATE() + INTERVAL 1 DAY").Where(limitBill).First(&limitBill).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			limit := custodyModels.Limit{
				UserId:    user.User.ID,
				LimitType: limitType.ID,
			}
			err = db.Where(limit).First(&limit).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {

				return nil, err
			}

			if errors.Is(err, gorm.ErrRecordNotFound) {

				limit.Level = defaultLimitLevel
				if err := db.Create(&limit).Error; err != nil {
					return nil, err
				}
			}

			if limit.Level == 0 {

				return nil, ErrLimitEntirely
			}

			levelLimit := custodyModels.LimitLevel{
				LimitTypeId: limitType.ID,
				Level:       limit.Level,
			}
			err = db.Where(levelLimit).First(&levelLimit).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {

				return nil, err
			}
			if errors.Is(err, gorm.ErrRecordNotFound) {

				return nil, nil
			}

			limitBill.TotalAmount = levelLimit.Amount
			limitBill.UseAbleAmount = levelLimit.Amount
			limitBill.TotalCount = levelLimit.Count
			limitBill.UseAbleCount = levelLimit.Count
			limitBill.LocalTime = time.Now().Add(-time.Minute * 15)

			if err := db.Create(&limitBill).Error; err != nil {
				return nil, err
			}

			return &limitBill, nil
		}

		return nil, err
	}

	return &limitBill, nil
}

func CheckLimit(db *gorm.DB, user *caccount.UserInfo, limitType *custodyModels.LimitType, amount float64) error {
	limitBill, err := GetLimit(db, user, limitType)
	if err != nil {
		return err
	}

	if limitBill == nil {

		return nil
	}

	if limitBill.UseAbleAmount < amount {
		return fmt.Errorf("%w,剩余额度：%v", errors.New("今日可用交易额度不足"), limitBill.UseAbleAmount)
	}
	if limitBill.UseAbleCount <= 0 {
		return fmt.Errorf("%w,剩余交易次数：%v", errors.New("今日可用交易次数不足"), limitBill.UseAbleCount)
	}
	sutime := time.Now().Sub(limitBill.LocalTime).Seconds()
	if sutime < 5 {
		return errors.New("交易频繁，请稍后再试")
	}
	return nil
}

func MinusLimit(db *gorm.DB, user *caccount.UserInfo, limitType *custodyModels.LimitType, amount float64) error {

	err := db.Where(limitType).First(limitType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			return nil
		}

		return err
	}

	limitMux.Lock()
	defer limitMux.Unlock()

	limitBill := custodyModels.LimitBill{
		UserId:    user.User.ID,
		LimitType: limitType.ID,
	}

	err = db.Where("created_at >= CURDATE() AND created_at < CURDATE() + INTERVAL 1 DAY").Where(limitBill).First(&limitBill).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			return nil
		}

		return err
	}

	if limitBill.UseAbleAmount < amount+20 {
		return errors.New("可用额度不足")
	}
	if limitBill.UseAbleCount <= 0 {
		return errors.New("可用交易次数不足")
	}
	sutime := time.Now().Sub(limitBill.LocalTime).Seconds()
	if sutime < 5 {
		return errors.New("交易频繁，请稍后再试")
	}

	limitBill.UseAbleAmount -= amount
	limitBill.UseAbleCount -= 1
	limitBill.LocalTime = time.Now()

	if err := db.Save(&limitBill).Error; err != nil {
		return err
	}

	return nil
}

func AddLimit(limitType int, userId int, limit int) {

}
