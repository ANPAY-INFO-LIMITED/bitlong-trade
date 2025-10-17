package services

import (
	"strconv"
	"trade/api"
	"trade/btlLog"
	"trade/models"
	"trade/services/assetsyncinfo"
	"trade/services/btldb"
)

func TradeToUserFundChannel(assetId, peerPubkey string, amount, feeRate, pushSat, localAmt int) (string, error) {
	if assetId == "00" {
		resp, err := api.OpenBtcChannel(amount, peerPubkey, feeRate, pushSat)
		if err != nil {
			btlLog.OpenChannel.Error("\nopen btc channel err\n%v", err)
			return "", err
		}
		return resp, nil
	}

	resp, err := api.FundChannel(assetId, amount, peerPubkey, feeRate, pushSat, localAmt)
	if err != nil {
		btlLog.OpenChannel.Error("\nopen asset channel err\n%v", err)
		return "", err
	}

	return resp, nil
}

func CreateOpenChannelRecord(openChan *models.AssetOrBtcChannelRecord) error {
	return btldb.CreateOpenChannelRecord(openChan)
}

func UpdateOpenChannelRecord(openChan *models.AssetOrBtcChannelRecord) error {
	_, err := btldb.ReadOpenChannelRecord(openChan.PaidId)
	if err != nil {
		return btldb.CreateOpenChannelRecord(openChan)
	}
	return btldb.UpdateOpenChannelRecord(openChan)
}

func CreateChannelAsset(asset *models.ChannelAssets) error {
	return btldb.CreateChannelAsset(asset)
}

func ReadChannelAsset(assetId string) (*models.ChannelAssets, error) {
	return btldb.ReadChannelAsset(assetId)
}

func UpdateChannelAsset(assetId string, updateData map[string]any) error {
	return btldb.UpdateChannelAsset(assetId, updateData)
}

func ValidateFundChannelRequest(assetID string, amount int, localSats int) string {
	decimal, err := assetsyncinfo.GetAssetsDecimal([]string{assetID})
	if err != nil {
		return "获取资产小数位失败"
	}
	asset, err := btldb.ReadChannelAsset(assetID)
	if err != nil || asset == nil {
		return "资产无权限"
	} else {
		if assetID == "00" {
			if amount < asset.MinSatsAmt || amount > asset.MaxSatsAmt || localSats != 0 {
				return "sats金额需大于" + strconv.Itoa(asset.MinSatsAmt) + "且小于" + strconv.Itoa(asset.MaxSatsAmt)
			}
			return ""
		} else {
			var decimalMultiplier int = 1
			if len(decimal) > 0 {
				for i := 0; i < int(decimal[0].DecimalDisplay); i++ {
					decimalMultiplier *= 10
				}
			}

			adjustedMinAmount := asset.MinAssetAmount * decimalMultiplier
			adjustedMaxAmount := asset.MaxAssetAmount * decimalMultiplier

			if amount < adjustedMinAmount || amount > adjustedMaxAmount {
				return "资产数量不满足条件"
			} else if localSats < asset.MinLocalSats || localSats > asset.MaxLocalSats {
				return "本地sats数量不满足条件"
			} else {
				return ""
			}
		}
	}
}
