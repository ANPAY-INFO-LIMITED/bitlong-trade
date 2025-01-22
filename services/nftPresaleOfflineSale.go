package services

import (
	"errors"
	"gorm.io/gorm"
	"trade/middleware"
	"trade/utils"
)

type NftPresaleOfflinePurchaseData struct {
	gorm.Model
	NftNo          string `json:"nft_no"`
	Name           string `json:"name"`
	NpubKey        string `json:"npub_key"`
	InvitationCode string `json:"invitation_code"`
	AssetId        string `json:"asset_id"`
	AssetName      string `json:"asset_name"`
	ReceiveAddr    string `json:"receive_addr"`
}

type NftPresaleOfflinePurchaseDataInfo struct {
	ID             uint   `json:"id"`
	NftNo          string `json:"nft_no"`
	NpubKey        string `json:"npub_key"`
	InvitationCode string `json:"invitation_code"`
	AssetId        string `json:"asset_id"`
	AssetName      string `json:"asset_name"`
}

func GetNftPresaleOfflinePurchaseData(nftNo string, npubKey string, invitationCode string, assetId string) (*[]NftPresaleOfflinePurchaseDataInfo, error) {
	var err error
	var where string
	if nftNo != "" {
		where += "nft_no = '" + nftNo + "'"
		if npubKey != "" {
			where += " AND npub_key = '" + npubKey + "'"
		}
		if invitationCode != "" {
			where += " AND invitation_code = '" + invitationCode + "'"
		}
		if assetId != "" {
			where += " AND asset_id = '" + assetId + "'"
		}
	} else {
		if npubKey != "" {
			where += "npub_key = '" + npubKey + "'"
			if invitationCode != "" {
				where += " AND invitation_code = '" + invitationCode + "'"
			}
			if assetId != "" {
				where += " AND asset_id = '" + assetId + "'"
			}
		} else {
			if invitationCode != "" {
				where += "invitation_code = '" + invitationCode + "'"
				if assetId != "" {
					where += " AND asset_id = '" + assetId + "'"
				}
			} else {
				if assetId != "" {
					where += "asset_id = '" + assetId + "'"
				} else {
					where = ""
				}
			}
		}
	}
	if where == "" {
		err = errors.New("no query condition")
		return new([]NftPresaleOfflinePurchaseDataInfo), err
	}
	var nftPresaleOfflinePurchaseDataInfos []NftPresaleOfflinePurchaseDataInfo
	err = middleware.DB.
		Table("nft_presale_offline_purchase_data").
		Select("id, nft_no, npub_key, invitation_code, asset_id, asset_name").
		Where(where).
		Scan(&nftPresaleOfflinePurchaseDataInfos).
		Error

	if err != nil {
		return new([]NftPresaleOfflinePurchaseDataInfo), utils.AppendErrorInfo(err, "select NftPresaleOfflinePurchaseData")
	}

	return &nftPresaleOfflinePurchaseDataInfos, nil
}

func UpdateNftPresaleOfflinePurchaseData(nftNo string, npubKey string, invitationCode string, name string) (err error) {
	return middleware.DB.
		Table("nft_presale_offline_purchase_data").
		Where("nft_no = ?", nftNo).
		Updates(map[string]any{
			"npub_key":        npubKey,
			"invitation_code": invitationCode,
			"name":            name,
		}).
		Error
}

type UpdateNftPresaleOfflinePurchaseDataRequest struct {
	NftNo          string `json:"nft_no"`
	NpubKey        string `json:"npub_key"`
	InvitationCode string `json:"invitation_code"`
	Name           string `json:"name"`
}
