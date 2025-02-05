package services

import (
	"trade/middleware"
	"trade/models"
)

type NftPresaleInfo struct {
	ID              uint                   `json:"id"`
	AssetId         string                 `json:"asset_id"`
	Name            string                 `json:"name"`
	Meta            string                 `json:"meta"`
	GroupKey        string                 `json:"group_key"`
	Price           int                    `json:"price"`
	Info            string                 `json:"info"`
	BuyerUsername   string                 `json:"buyer_username"`
	ReceiveAddr     string                 `json:"receive_addr"`
	BoughtTime      int                    `json:"bought_time"`
	PaidId          int                    `json:"paid_id"`
	PaidSuccessTime int                    `json:"paid_success_time"`
	State           models.NftPresaleState `json:"state"`
}

func GetPurchasedNftPresaleInfo() ([]NftPresaleInfo, error) {
	db := middleware.DB
	var nftPresaleInfos []NftPresaleInfo
	err := db.Table("nft_presales").
		Select("id, asset_id, name, meta, group_key, price, info, buyer_username, receive_addr, bought_time, paid_id, paid_success_time, state").
		Where("state > ?", models.NftPresaleStatePaidPending).
		Order("bought_time desc").
		Scan(&nftPresaleInfos).
		Error
	if err != nil {
		return nil, err
	}
	return nftPresaleInfos, nil
}

func GetPurchasedNftPresaleInfoCount() (count int64, err error) {
	db := middleware.DB
	var nftPresaleInfos []NftPresaleInfo
	err = db.Table("nft_presales").
		Where("state > ?", models.NftPresaleStatePaidPending).
		Count(&count).
		Scan(&nftPresaleInfos).
		Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetPurchasedNftPresaleInfoLimitAndOffset(limit int, offset int) ([]NftPresaleInfo, error) {
	db := middleware.DB
	var nftPresaleInfos []NftPresaleInfo
	err := db.Table("nft_presales").
		Select("id, asset_id, name, meta, group_key, price, info, buyer_username, receive_addr, bought_time, paid_id, paid_success_time, state").
		Where("state > ?", models.NftPresaleStatePaidPending).
		Order("bought_time desc").
		Limit(limit).
		Offset(offset).
		Scan(&nftPresaleInfos).
		Error
	if err != nil {
		return nil, err
	}
	return nftPresaleInfos, nil
}
