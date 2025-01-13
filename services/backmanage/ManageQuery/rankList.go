package ManageQuery

import (
	"sort"
	"trade/services/custodyAccount/localQuery"
	"trade/services/pool"
)

type GetAssetRankListResp struct {
	AssetId  string  `json:"assetId"`
	UserName string  `json:"userName"`
	Amount   float64 `json:"amount"`
}

func GetAssetsBalanceRankList(assetId string, pageNum int, pageSize int) (*[]GetAssetRankListResp, int64, float64) {
	var resps []GetAssetRankListResp
	var totalAmount float64
	var totalCount int64

	quest := localQuery.GetAssetListQuest{
		AssetId:  assetId,
		Page:     pageNum,
		PageSize: pageSize,
	}
	r, c, t := localQuery.GetAssetList(quest)
	totalAmount += t
	totalCount += c
	for _, v := range *r {
		resps = append(resps, GetAssetRankListResp{
			AssetId:  v.AssetId,
			UserName: v.UserName,
			Amount:   v.Amount,
		})
	}

	count, err := pool.GetPoolAccountNameAndBalancesCount(assetId)
	if err != nil {
		return nil, 0, 0
	}
	totalCount += count

	pools, err := pool.GetPoolAccountTotalBalance(assetId)
	if err != nil {
		return nil, 0, 0
	}
	totalAmount += pools

	balances, err := pool.GetPoolAccountNameAndBalances(assetId, pageSize, pageNum)
	if err != nil {
		return nil, 0, 0
	}
	for _, v := range balances {
		resps = append(resps, GetAssetRankListResp{
			AssetId:  assetId,
			UserName: v.Name,
			Amount:   v.Balance,
		})
	}
	sort.Slice(resps, func(i, j int) bool {
		return resps[i].Amount > resps[j].Amount
	})
	list := make([]GetAssetRankListResp, 0)
	if len(resps) > pageSize {
		list = resps[0:pageSize]
	} else {
		list = resps
	}
	return &list, totalCount, totalAmount
}
