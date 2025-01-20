package assetMoreInfo

import (
	"trade/middleware"
)

func GetAssetBurnTotal(assetId string) (assetBurnTotalAmount int64, err error) {
	err = middleware.DB.Table("asset_burns").
		Select("sum(amount) as total").
		Where("asset_id = ?", assetId).
		Scan(&assetBurnTotalAmount).Error
	return assetBurnTotalAmount, nil
}
