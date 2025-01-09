package Query

type GetAssetRankListResp struct {
	AssetId  string           `json:"assetId"`
	UserName string           `json:"userName"`
	Amount   float64          `json:"amount"`
	Details  []BalanceDetails `json:"balanceDetails"`
}
type BalanceDetails struct {
	BalanceType string  `json:"balanceType"`
	Amount      float64 `json:"amount"`
}

func GetAssetsBalanceRankList() (*[]GetAssetRankListResp, error) {

	return nil, nil
}
