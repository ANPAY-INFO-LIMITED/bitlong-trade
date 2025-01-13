package routers

import (
	"github.com/gin-gonic/gin"
	"trade/config"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	if !config.GetLoadConfig().RouterDisable.Login {
		SetupLoginRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.FairLaunch {
		SetupFairLaunchRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.Fee {
		SetupFeeRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.CustodyAccount {
		SetupCustodyAccountRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.Proof {
		SetupProofRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.Ido {
		SetupIdoRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.Snapshot {
		SetupSnapshotRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.BtcBalance {
		SetupBtcBalanceRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetTransfer {
		SetupAssetTransferRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.Bitcoind {
		SetupBitcoindRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.Shell {
		SetupShellRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AddrReceive {
		SetupAddrReceiveRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.BatchTransfer {
		SetupBatchTransferRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetAddr {
		SetupAssetAddrRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetLock {
		SetupAssetLockRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.ValidateToken {
		SetupValidateTokenRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetBalance {
		SetupAssetBalanceRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetBurn {
		SetupAssetBurnRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetLocalMint {
		SetupAssetLocalMintRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.Ping {
		SetupPingRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.User {
		SetupUserRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetRecommend {
		SetupAssetRecommendRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.FairLaunchFollow {
		SetupFairLaunchFollowRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetLocalMintHistory {
		SetupAssetLocalMintHistoryRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetManagedUtxo {
		SetupAssetManagedUtxoRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.LogFileUpload {
		SetupLogFileUploadRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AccountAsset {
		SetupAccountAssetRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetGroup {
		SetupAssetGroupRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.NftTransfer {
		SetupNftTransferRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.NftInfo {
		SetupNftInfoRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.NftPresale {
		SetupNftPresaleRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.UserData {
		SetupUserDataRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetList {
		SetupAssetListRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.UserStats {
		SetupUserStatsRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetInfo {
		SetupAssetInfoRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.BackReward {
		SetupBackRewardRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.Download {
		SetupDownloadRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetBalanceBackup {
		SetupAssetBalanceBackupRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetBalanceHistory {
		SetupAssetBalanceHistoryRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetMeta {
		SetupAssetMetaRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.Pool {
		SetupPoolRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.BtcUtxo {
		SetupBtcUtxoRouter(r)
	}
	if !config.GetLoadConfig().RouterDisable.AssetBalanceBackend {
		SetupAssetBalanceBackendRouter(r)
	}
	SetupWsRouter(r)
	return r
}
