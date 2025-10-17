package custodyAssets

import (
	"encoding/hex"
	"time"
	"trade/btlLog"
	"trade/middleware"
	"trade/models"
	"trade/models/custodyModels"
	"trade/services/btldb"
	rpc "trade/services/servicesrpc"
)

var isAble = true

func GoOutsideMission() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {

			if isAble {
				isAble = !isAble
				startOutsideMission()
				isAble = !isAble
			}
			<-ticker.C
		}
	}()
}

func startOutsideMission() {
	var results []struct {
		AssetID      string    `gorm:"column:asset_id"`
		MinCreatedAt time.Time `gorm:"column:min_created_at"`
	}
	db := middleware.DB
	db.Table("user_account_outside_asset_mission").
		Select("asset_id, MIN(created_at) as min_created_at").
		Where("status = ?", 0).
		Group("asset_id").
		Order("min_created_at").
		Scan(&results)
	if len(results) == 0 {
		return
	}

	assets, err := rpc.ListAssets()
	if err != nil {
		return
	}
	list := make(map[string]uint64)
	for _, asset := range assets.Assets {
		assetId := hex.EncodeToString(asset.AssetGenesis.AssetId)
		list[assetId] += asset.Amount
	}

	for _, result := range results {
		if list[result.AssetID] == 0 {
			continue
		}
		var outsideMissions []custodyModels.PayOutside
		db.Where("asset_id =? and status =?", result.AssetID, custodyModels.PayOutsideStatusPending).Limit(8).Find(&outsideMissions)
		if outsideMissions == nil || len(outsideMissions) == 0 {
			continue
		}

		missions := removeDuplicates(outsideMissions, list)
		if len(missions) == 0 {
			continue
		}

		balance, err := rpc.GetBalance()
		if err != nil || balance.AccountBalance["default"].ConfirmedBalance < int64(len(missions)*1000) {
			continue
		}

		payToOutside(&missions)
	}
}
func payToOutside(missions *[]custodyModels.PayOutside) {
	tx, back := middleware.GetTx()

	defer back()
	var err error

	var addr []string
	var balances []*models.Balance
	for index := range *missions {

		(*missions)[index].Status = custodyModels.PayOutsideStatusPaid
		err = btldb.UpdatePayOutside(tx, &(*missions)[index])
		if err != nil {
			btlLog.CUST.Error("btldb.UpdatePayOutside error:%w", err)
			return
		}

		balance, err := btldb.ReadBalance((*missions)[index].BalanceId)
		if err != nil {
			return
		}
		balances = append(balances, balance)
		balance.State = models.STATE_SUCCESS

		err = btldb.UpdateBalance(tx, balance)
		if err != nil {
			btlLog.CUST.Error("payToOutsideOnChain db error")
			return
		}
		addr = append(addr, (*missions)[index].Address)
	}
	response, err := rpc.SendAssets(addr)
	if err != nil {
		btlLog.CUST.Error("rpc.SendAssets error:%v,%v", err, addr)
		return
	}
	tx.Commit()

	b := response.Transfer.AnchorTxHash
	for i := 0; i < len(b)/2; i++ {
		temp := b[i]
		b[i] = b[len(b)-i-1]
		b[len(b)-i-1] = temp
	}
	txId := hex.EncodeToString(b)
	btctx := custodyModels.PayOutsideTx{
		TxHash:     txId,
		Timestamp:  response.Transfer.TransferTimestamp,
		HeightHint: response.Transfer.AnchorTxHeightHint,
		ChainFees:  response.Transfer.AnchorTxChainFees,
		InputsNum:  uint(len(response.Transfer.Inputs)),
		OutputsNum: uint(len(response.Transfer.Outputs)),
		Status:     custodyModels.PayOutsideStatusTXPending,
	}
	err = btldb.CreatePayOutsideTx(&btctx)
	if err != nil {
		btlLog.CUST.Error("btldb.CreatePayOutsideTx error:%w", err)
	}
	db := middleware.DB
	for index := range *missions {
		(*missions)[index].TxHash = txId
		err = btldb.UpdatePayOutside(db, &(*missions)[index])
		if err != nil {
			btlLog.CUST.Error("btldb.UpdatePayOutside error:%w", err)
		}
	}
	for index := range balances {
		balances[index].PaymentHash = &txId
		err = btldb.UpdateBalance(db, balances[index])
		if err != nil {
			btlLog.CUST.Error("payToOutsideOnChain db error")
		}
	}
}

func removeDuplicates(outsideMissions []custodyModels.PayOutside, list map[string]uint64) []custodyModels.PayOutside {

	unique := make(map[string]custodyModels.PayOutside)
	amount := uint64(0)

	for _, outsideMission := range outsideMissions {
		if _, exist := unique[outsideMission.Address]; !exist {
			amount += uint64(outsideMission.Amount)
			if amount > list[outsideMission.AssetId] {
				break
			}
			unique[outsideMission.Address] = outsideMission
		}
	}

	result := make([]custodyModels.PayOutside, 0, len(unique))
	for _, outside := range unique {
		result = append(result, outside)
	}

	return result
}
