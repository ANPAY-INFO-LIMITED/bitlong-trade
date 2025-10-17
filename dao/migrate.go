package dao

import (
	"trade/middleware"
	"trade/models"
	"trade/models/custodyModels"
	"trade/models/custodyModels/custodyswap"
	"trade/models/custodyModels/game"
	"trade/models/custodyModels/pAccount"
	"trade/models/custodyModels/store"
	"trade/models/forwardtrans"
	"trade/models/nodemanage"
	"trade/services"
	"trade/services/lntOfficial"
	"trade/services/pool"
	"trade/services/satBackQueue"
)

func Migrate() error {
	var err error
	defaultMigrate := []interface{}{
		&models.Balance{},
		&models.User{},
		&models.UserConfig{},
		&models.ScheduledTask{},
		&models.Invoice{},
		&models.InvoiceRfqInfo{},
		&models.FairLaunchInfo{},
		&models.FairLaunchMintedInfo{},
		&models.FairLaunchMintedUserInfo{},
		&models.FeeRateInfo{},
		&models.AssetIssuance{},
		&models.IdoPublishInfo{},
		&models.IdoParticipateInfo{},
		&models.IdoParticipateUserInfo{},
		&models.AssetSyncInfo{},
		&models.BtcBalance{},
		&models.AssetTransferProcessedDb{},
		&models.AssetTransferProcessedInputDb{},
		&models.AssetTransferProcessedOutputDb{},
		&models.AddrReceiveEvent{},
		&models.BatchTransfer{},
		&models.AssetAddr{},
		&models.AssetLock{},
		&models.AssetBalance{},
		&models.AssetBurn{},
		&models.AssetLocalMint{},
		&models.AssetRecommend{},
		&models.LoginRecord{},
		&models.FairLaunchFollow{},
		&models.AssetLocalMintHistory{},
		&models.AssetManagedUtxo{},
		&models.FairLaunchMintedAndAvailableInfo{},
		&models.FairLaunchIncome{},
		&models.BackFee{},
		&models.LogFileUpload{},
		&models.AccountAssetReceive{},
		&models.AssetGroup{},
		&models.NftTransfer{},
		&models.NftInfo{},
		&models.NftPresale{},
		&models.AssetMeta{},
		&models.NftPresaleBatchGroup{},
		&models.NftPresaleWhitelist{},
		&models.AssetList{},
		&models.DateIpLogin{},
		&models.DateLogin{},
		&models.BalanceTypeExt{},
		&models.AssetBalanceBackup{},
		&models.AssetBalanceHistory{},
		&satBackQueue.PushQueueRecord{},
		&satBackQueue.SwapTrPushQueueRecord{},
		&models.RestRecord{},
		&services.NftPresaleOfflinePurchaseData{},
		&models.BtcUtxo{},
		&models.BtcUtxoHistory{},
		&models.PsbtTlSwap{},
		&pool.PoolPair{},
		&pool.PoolShare{},
		&pool.PoolShareBalance{},
		&pool.PoolShareRecord{},
		&pool.PoolSwapRecord{},
		&pool.PoolLpAwardBalance{},
		&pool.PoolLpAwardRecord{},
		&pool.PoolWithdrawAwardRecord{},
		&pool.PoolAddLiquidityBatch{},
		&pool.PoolRemoveLiquidityBatch{},
		&pool.PoolSwapExactTokenForTokenNoPathBatch{},
		&pool.PoolSwapTokenForExactTokenNoPathBatch{},
		&pool.PoolWithdrawAwardBatch{},
		&pool.PoolAccountFeeBalance{},
		&pool.PoolPairTokenAccountBalance{},
		&pool.PoolLpAwardCumulative{},
		&pool.PoolShareLpAwardBalance{},
		&pool.PoolShareLpAwardCumulative{},
		&pool.PoolPureAddLiquidityRecord{},
		&pool.PoolBeforeSwapFee{},
		&satBackQueue.GenLiquidity{},
		&satBackQueue.GenLiquidityPushQueueRecord{},
		&models.LitConf{},
		&models.BoxDevice{},
		&models.BoxDeviceBackup{},
		&models.BoxChannelsInfo{},
		&models.TradeChannelsInfo{},
		&models.BoxAssetPush{},
		&models.BoxAssetPushTask{},
		&models.BoxAutoAssetChannelInfo{},
		&models.BoxIP{},
		&models.BoxFrp{},
		&models.BoxFrpHistory{},
		&models.BoxProxyReqHistory{},
		&models.ChannelSwapMission{},
		&nodemanage.LitNodeConfig{},
		&models.LoDataHistory{},
		&models.DeliAddr{},
		&models.DeliAddrRecInfo{},
		lntOfficial.ContactInfo{},
		&forwardtrans.FwdTransInvoiceMapping{},
		&forwardtrans.ToCustody{},
	}
	if err = middleware.DB.AutoMigrate(defaultMigrate...); err != nil {
		return err
	}
	{
		if err = assetOrBtcChannelMigrate(err); err != nil {
			return err
		}
		if err = custody(err); err != nil {
			return err
		}
	}
	return err
}

func assetOrBtcChannelMigrate(err error) error {
	m := []interface{}{
		&models.AssetChannelNode{},
		&models.AssetOrBtcChannelRecord{},
		&models.ChannelAssets{},
	}
	if err = middleware.DB.AutoMigrate(m...); err != nil {
		return err
	}
	return err
}
func custody(err error) error {
	m := []interface{}{
		&models.Account{},
		&store.Store{},
		&store.StoreInfo{},
		&store.PendingPropose{},
		&custodyModels.LockBill{},
		&custodyModels.LockAccount{},
		&custodyModels.LockBalance{},
		&custodyModels.LockBillExt{},
		&models.AccountAwardExt{},
		&models.AwardInventory{},
		&models.AccountAward{},
		&models.AccountAwardIdempotent{},
		&custodyModels.Limit{},
		&custodyModels.LimitBill{},
		&custodyModels.LimitLevel{},
		&custodyModels.LimitType{},
		&custodyModels.BlockedRecord{},
		&custodyModels.AssetFee{},
		&custodyswap.ReceiveConfig{},
		&custodyswap.SwapAwardRecord{},
		&custodyswap.SwapSupplier{},
		&custodyswap.SwapBill{},
		&custodyModels.Control{},
		&custodyModels.AccountInsideMission{},
		&custodyModels.AccountOutsideMission{},
		&custodyModels.AccountBalanceChange{},
		custodyModels.OutBtcOnChain{},
		custodyModels.RechargeBtcOnChain{},
		&custodyModels.AccountBtcBalance{},
		&custodyModels.AwardConf{},
		&pAccount.PoolAccount{},
		&pAccount.PAccountAssetId{},
		&pAccount.PAccountBalance{},
		&pAccount.PAccountBill{},
		&pAccount.PAccountBalanceChange{},
		&custodyModels.AccountBalance{},
		&custodyModels.PayOutside{},
		&custodyModels.PayOutsideTx{},
		&store.ReviewAward{},
		&game.Recharge{},
		&game.Withdraw{},
	}
	if err = middleware.DB.AutoMigrate(m...); err != nil {
		return err
	}
	return err
}
