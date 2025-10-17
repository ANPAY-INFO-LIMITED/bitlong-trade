package localQuery

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"trade/middleware"
	"trade/models/custodyModels"
	"trade/services/pool"
)

type ChannelQueryQuest struct {
	AssetId string `json:"assetId"`
}
type ChannelQueryResp struct {
	TotalAmount float64 `json:"totalAmount"`
}

var DbError = errors.New("db error")

func QueryChannelAssetInfo(quest *ChannelQueryQuest) (*ChannelQueryResp, error) {
	db := middleware.DB

	var total float64
	if quest.AssetId == "00" {
		err := db.Model(&custodyModels.AccountBtcBalance{}).Select("SUM(amount) as total").Scan(&total).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, DbError
		}
	} else {
		q := db.Where("asset_id =?", quest.AssetId)

		err := q.Model(&custodyModels.AccountBalance{}).Select("SUM(amount) as total").Scan(&total).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, DbError
		}
	}

	var resp ChannelQueryResp
	resp.TotalAmount = total
	return &resp, nil
}

func QueryChannelAssetWithOutAdmin(quest *ChannelQueryQuest) (float64, error) {
	var err error
	db := middleware.DB

	ct := struct {
		Count int64   `gorm:"column:count"`
		Total float64 `gorm:"column:total"`
	}{}

	if quest.AssetId == "" {
		return 0, fmt.Errorf("assetId is empty")
	} else if quest.AssetId != "00" {
		err = db.Raw(assetListQueryCT, quest.AssetId, quest.AssetId, quest.AssetId).Scan(&ct).Error
		if err != nil {
			return 0, err
		}
	} else {
		err = db.Raw(btcListQueryCT).Scan(&ct).Error
		if err != nil {
			return 0, err
		}
	}
	pools, err := pool.GetPoolAccountTotalBalance(quest.AssetId)
	if err != nil {
		return 0, err
	}
	ct.Total += pools

	var withoutTotal float64
	if quest.AssetId == "00" {
		err := db.Raw(WithOutAdminSqlBtc).Scan(&withoutTotal).Error
		if err != nil {
			return 0, err
		}
	} else {
		err := db.Raw(WithOutAdminSql, quest.AssetId).Scan(&withoutTotal).Error
		if err != nil {
			return 0, err
		}
	}
	return ct.Total - withoutTotal, nil
}

var WithOutAdminSql = `
select COALESCE(sum(amount),0)
from user_account_balance
where account_id IN (
    SELECT id
    FROM user_account
    WHERE user_name IN ('admin','blackhole')
    )and asset_id = ?
`
var WithOutAdminSqlBtc = `
select COALESCE(sum(amount),0)
from user_account_balance_btc
where account_id IN (
    SELECT id
    FROM user_account
    WHERE user_name IN ('admin','blackhole')
    )
`
