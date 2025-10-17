package services

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
	"trade/api"
	"trade/btlLog"
	"trade/config"
	"trade/middleware"
	"trade/models"
	"trade/services/lntOfficial"
	"trade/services/pool"
	"trade/services/psbtTlSwap"
	"trade/services/satBackQueue"
	"trade/utils"
)

type CronService struct{}

func CheckIfAutoUpdateScheduledTask() {
	if config.GetLoadConfig().IsAutoUpdateScheduledTask {
		err := CreateFairLaunchProcessions()
		if err != nil {
			btlLog.ScheduledTask.Info("%v", err)
		}
		err = CreateSnapshotProcessions()
		if err != nil {
			btlLog.ScheduledTask.Info("%v", err)
		}
		err = CreateSetTransfersAndReceives()
		if err != nil {
			btlLog.ScheduledTask.Info("%v", err)
		}
		err = CreateNftPresaleProcessions()
		if err != nil {
			btlLog.ScheduledTask.Info("%v", err)
		}
		err = CreatePushQueueProcessions()
		if err != nil {
			btlLog.ScheduledTask.Info("%v", err)
		}
		err = CreatePoolPairTokenAccountBalanceProcessions()
		if err != nil {
			btlLog.ScheduledTask.Info("%v", err)
		}
		err = CreatePsbtTlSwapProcessPendingTx()
		if err != nil {
			btlLog.ScheduledTask.Info("%v", err)
		}
		err = CreateBoxAssetPushProcessions()
		if err != nil {
			btlLog.ScheduledTask.Info("%v", err)
		}
		err = CreateLoDataHistory()
		if err != nil {
			btlLog.ScheduledTask.Info("%v", err)
		}
	}
}

func CreateFairLaunchProcessions() (err error) {
	return CreateOrUpdateScheduledTasks(&[]models.ScheduledTask{
		{
			Name:           "ProcessFairLaunchNoPay",
			CronExpression: "*/20 * * * * *",
			FunctionName:   "ProcessFairLaunchNoPay",
			Package:        "services",
		}, {
			Name:           "ProcessFairLaunchPaidPending",
			CronExpression: "*/20 * * * * *",
			FunctionName:   "ProcessFairLaunchPaidPending",
			Package:        "services",
		}, {
			Name:           "ProcessFairLaunchPaidNoIssue",
			CronExpression: "0 */2 * * * *",
			FunctionName:   "ProcessFairLaunchPaidNoIssue",
			Package:        "services",
		}, {
			Name:           "ProcessFairLaunchIssuedPending",
			CronExpression: "*/20 * * * * *",
			FunctionName:   "ProcessFairLaunchIssuedPending",
			Package:        "services",
		}, {
			Name:           "ProcessFairLaunchReservedSentPending",
			CronExpression: "*/20 * * * * *",
			FunctionName:   "ProcessFairLaunchReservedSentPending",
			Package:        "services",
		}, {
			Name:           "FairLaunchMint",
			CronExpression: "*/20 * * * * *",
			FunctionName:   "FairLaunchMint",
			Package:        "services",
		}, {
			Name:           "FairLaunchMintSentPendingCheck",
			CronExpression: "0 */3 * * * *",
			FunctionName:   "FairLaunchMintSentPendingCheck",
			Package:        "services",
		}, {
			Name:           "SendFairLaunchAsset",
			CronExpression: "0 */5 * * * *",
			FunctionName:   "SendFairLaunchAsset",
			Package:        "services",
		},
		{
			Name:           "SnapshotToZipLast",
			CronExpression: "0 */5 * * * *",
			FunctionName:   "SnapshotToZipLast",
			Package:        "services",
		},
		{
			Name:           "UpdateFairLaunchIncomesSatAmountByTxids",
			CronExpression: "*/20 * * * * *",
			FunctionName:   "UpdateFairLaunchIncomesSatAmountByTxids",
			Package:        "services",
		},
	})
}

func TaskCountRecordByRedis(name string) error {
	var record string
	var count int
	var err error
	record, err = middleware.RedisGet(name)
	if err != nil {

		err = middleware.RedisSet(name, "1"+","+utils.GetTimeNow(), 6*time.Minute)
		if err != nil {
			return err
		}
		return nil
	}
	split := strings.Split(record, ",")
	count, err = strconv.Atoi(split[0])
	if err != nil {
		return err
	}
	err = middleware.RedisSet(name, strconv.Itoa(count+1)+","+utils.GetTimeNow(), 6*time.Minute)
	if err != nil {
		return err
	}
	return nil
}

func (cs *CronService) FairLaunchIssuance() {
	tx := middleware.DB.Begin()
	FairLaunchIssuance(tx)
	err := TaskCountRecordByRedis("FairLaunchIssuance")
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (cs *CronService) SnapshotToZipLast() {
	SnapshotToZipLast()
	err := TaskCountRecordByRedis("SnapshotToZipLast")
	if err != nil {
		return
	}
}

func (cs *CronService) ProcessFairLaunchNoPay() {
	tx := middleware.DB.Begin()
	ProcessFairLaunchNoPay(tx)
	err := TaskCountRecordByRedis("ProcessFairLaunchNoPay")
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (cs *CronService) ProcessFairLaunchPaidPending() {
	tx := middleware.DB.Begin()
	ProcessFairLaunchPaidPending(tx)
	err := TaskCountRecordByRedis("ProcessFairLaunchPaidPending")
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (cs *CronService) ProcessFairLaunchPaidNoIssue() {
	tx := middleware.DB.Begin()
	ProcessFairLaunchPaidNoIssue(tx)
	err := TaskCountRecordByRedis("ProcessFairLaunchPaidNoIssue")
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (cs *CronService) ProcessFairLaunchIssuedPending() {
	tx := middleware.DB.Begin()
	ProcessFairLaunchIssuedPending(tx)
	err := TaskCountRecordByRedis("ProcessFairLaunchIssuedPending")
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (cs *CronService) ProcessFairLaunchReservedSentPending() {
	tx := middleware.DB.Begin()
	ProcessFairLaunchReservedSentPending(tx)
	err := TaskCountRecordByRedis("ProcessFairLaunchReservedSentPending")
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (cs *CronService) FairLaunchMint() {
	tx := middleware.DB.Begin()
	FairLaunchMint(tx)
	err := TaskCountRecordByRedis("FairLaunchMint")
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (cs *CronService) FairLaunchMintSentPendingCheck() {
	tx := middleware.DB.Begin()
	FairLaunchMintSentPendingCheck(tx)
	err := TaskCountRecordByRedis("FairLaunchMintSentPendingCheck")
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (cs *CronService) SendFairLaunchAsset() {

	SendFairLaunchAsset()
	err := TaskCountRecordByRedis("SendFairLaunchAsset")
	if err != nil {

		return
	}

}

func (cs *CronService) RemoveMintedInventories() {
	RemoveMintedInventories()
	err := TaskCountRecordByRedis("RemoveMintedInventories")
	if err != nil {
		return
	}
}

func CreateSnapshotProcessions() (err error) {
	return CreateOrUpdateScheduledTasks(&[]models.ScheduledTask{
		{
			Name:           "SnapshotToZipLast",
			CronExpression: "0 0 */12 * * *",
			FunctionName:   "SnapshotToZipLast",
			Package:        "services",
		},
	})
}

func (cs *CronService) ListAndSetAssetTransfers() {
	network, err := api.NetworkStringToNetwork(config.GetLoadConfig().NetWork)
	if err != nil {
		return
	}
	userByte := sha256.Sum256([]byte(AdminUploadUserName))
	username := hex.EncodeToString(userByte[:])
	if len(username) < 16 {
		return
	}
	deviceId := username[:16]
	err = ListAndSetAssetTransfers(network, deviceId)
	if err != nil {
		return
	}
}

func (cs *CronService) GetAndSetAddrReceivesEvents() {
	userByte := sha256.Sum256([]byte(AdminUploadUserName))
	username := hex.EncodeToString(userByte[:])
	if len(username) < 16 {
		return
	}
	deviceId := username[:16]
	err := GetAndSetAddrReceivesEvents(deviceId)
	if err != nil {
		return
	}
}

func CreateSetTransfersAndReceives() (err error) {
	return CreateOrUpdateScheduledTasks(&[]models.ScheduledTask{
		{
			Name:           "ListAndSetAssetTransfers",
			CronExpression: "0 */5 * * * *",
			FunctionName:   "ListAndSetAssetTransfers",
			Package:        "services",
		},
		{
			Name:           "GetAndSetAddrReceivesEvents",
			CronExpression: "0 */3 * * * *",
			FunctionName:   "GetAndSetAddrReceivesEvents",
			Package:        "services",
		},
	})
}

func (cs *CronService) UpdateFairLaunchIncomesSatAmountByTxids() {
	network, err := api.NetworkStringToNetwork(config.GetLoadConfig().NetWork)
	if err != nil {
		return
	}
	err = UpdateFairLaunchIncomesSatAmountByTxids(network)
	if err != nil {
		return
	}
}

func (cs *CronService) ProcessNftPresaleBoughtNotPay() {
	ProcessNftPresaleBoughtNotPay()
	err := TaskCountRecordByRedis("ProcessNftPresaleBoughtNotPay")
	if err != nil {
		return
	}
}

func (cs *CronService) ProcessNftPresalePaidPending() {
	ProcessNftPresalePaidPending()
	err := TaskCountRecordByRedis("ProcessNftPresalePaidPending")
	if err != nil {
		return
	}
}

func (cs *CronService) ProcessNftPresalePaidNotSend() {
	ProcessNftPresalePaidNotSend()
	err := TaskCountRecordByRedis("ProcessNftPresalePaidNotSend")
	if err != nil {
		return
	}
}

func (cs *CronService) ProcessNftPresaleSentPending() {
	ProcessNftPresaleSentPending()
	err := TaskCountRecordByRedis("ProcessNftPresaleSentPending")
	if err != nil {
		return
	}
}

func CreateNftPresaleProcessions() (err error) {
	return CreateOrUpdateScheduledTasks(&[]models.ScheduledTask{
		{
			Name:           "ProcessNftPresaleBoughtNotPay",
			CronExpression: "*/20 * * * * *",
			FunctionName:   "ProcessNftPresaleBoughtNotPay",
			Package:        "services",
		},
		{
			Name:           "ProcessNftPresalePaidPending",
			CronExpression: "*/20 * * * * *",
			FunctionName:   "ProcessNftPresalePaidPending",
			Package:        "services",
		},
		{
			Name:           "ProcessNftPresalePaidNotSend",
			CronExpression: "0 */5 * * * *",
			FunctionName:   "ProcessNftPresalePaidNotSend",
			Package:        "services",
		},
		{
			Name:           "ProcessNftPresaleSentPending",
			CronExpression: "*/20 * * * * *",
			FunctionName:   "ProcessNftPresaleSentPending",
			Package:        "services",
		},
	})
}

func CreatePushQueueProcessions() (err error) {
	return CreateOrUpdateScheduledTasks(&[]models.ScheduledTask{
		{
			Name:           "GetAndPushClaimAsset",
			CronExpression: "*/30 * * * * *",
			FunctionName:   "GetAndPushClaimAsset",
			Package:        "services",
		},
		{
			Name:           "GetAndPushPurchasePresaleNFT",
			CronExpression: "*/30 * * * * *",
			FunctionName:   "GetAndPushPurchasePresaleNFT",
			Package:        "services",
		},
		{
			Name:           "GetAndPushSwapTrs",
			CronExpression: "*/30 * * * * *",
			FunctionName:   "GetAndPushSwapTrs",
			Package:        "services",
		},
		{
			Name:           "GetAndPushGenLiquidity",
			CronExpression: "*/30 * * * * *",
			FunctionName:   "GetAndPushGenLiquidity",
			Package:        "services",
		},
	})
}

func (cs *CronService) GetAndPushClaimAsset() {
	satBackQueue.GetAndPushClaimAsset()
}

func (cs *CronService) GetAndPushPurchasePresaleNFT() {
	satBackQueue.GetAndPushPurchasePresaleNFT()
}

func (cs *CronService) GetAndPushSwapTrs() {
	satBackQueue.GetAndPushSwapTrs()
}

func (cs *CronService) GetAndPushGenLiquidity() {
	satBackQueue.GetAndPushGenLiquidity()
}

func CreatePoolPairTokenAccountBalanceProcessions() (err error) {
	return CreateOrUpdateScheduledTasks(&[]models.ScheduledTask{
		{
			Name:           "UpdatePoolPairTokenAccountBalance",
			CronExpression: "0 */2 * * * *",
			FunctionName:   "UpdatePoolPairTokenAccountBalance",
			Package:        "services",
		},
	})
}

func (cs *CronService) UpdatePoolPairTokenAccountBalance() {
	err := pool.UpdateAllPoolPairTokenAccountBalances()
	if err != nil {
		btlLog.PoolPairTokenAccountBalance.Error("%v", err)
		return
	}
}

func CreatePsbtTlSwapProcessPendingTx() (err error) {
	return CreateOrUpdateScheduledTasks(&[]models.ScheduledTask{
		{
			Name:           "ProcessPendingTx",
			CronExpression: "0 */2 * * * *",
			FunctionName:   "ProcessPendingTx",
			Package:        "services",
		},
	})
}

func (cs *CronService) ProcessPendingTx() {
	err := psbtTlSwap.ProcessPendingTx()
	if err != nil {
		btlLog.PsbtTlSwap.Error("ProcessPendingTx error: %v", err)
		return
	}
}

func CreateBoxAssetPushProcessions() (err error) {
	return CreateOrUpdateScheduledTasks(&[]models.ScheduledTask{
		{
			Name:           "OpenBoxAssetChannels",
			CronExpression: "0 */8 * * * *",
			FunctionName:   "OpenBoxAssetChannels",
			Package:        "services",
		},
		{
			Name:           "PushBoxAsset",
			CronExpression: "0 */10 * * * *",
			FunctionName:   "PushBoxAsset",
			Package:        "services",
		},
		{
			Name:           "SetTradeChannelsInfo",
			CronExpression: "0 */2 * * * *",
			FunctionName:   "SetTradeChannelsInfo",
			Package:        "services",
		},
	})
}

func (cs *CronService) OpenBoxAssetChannels() {
	btlLog.BoxAssetPush.Info("OpenBoxAssetChannelsStart")
	cfg := config.GetLoadConfig()
	if cfg.NetWork != "mainnet" {
		return
	}
	err := OpenBoxAssetChannels()
	if err != nil {
		btlLog.BoxAssetPush.Error("EndOpenBoxAssetChannels error: %v", err)
		return
	}
	btlLog.BoxAssetPush.Info("EndOpenBoxAssetChannels")
}

func (cs *CronService) PushBoxAsset() {
	btlLog.BoxAssetPush.Info("PushBoxAssetStart")
	cfg := config.GetLoadConfig()
	if cfg.NetWork != "mainnet" {
		return
	}
	startTime := time.Now()

	err := PushBoxAsset()
	if err != nil {
		btlLog.BoxAssetPush.Error("EndPushBoxAsset error: %v, 耗时: %v", err, time.Since(startTime))
		return
	}
	btlLog.BoxAssetPush.Info("EndPushBoxAsset - 执行成功，耗时: %v", time.Since(startTime))
}

func (cs *CronService) SetTradeChannelsInfo() {
	cfg := config.GetLoadConfig()
	if cfg.NetWork != "mainnet" {
		return
	}
	btlLog.BoxChannelInfos.Info("SetTradeChannelsInfoStart")
	err := SetTradeChannelsInfo()
	if err != nil {
		btlLog.BoxChannelInfos.Error("SetTradeChannelsInfo error: %v", err)
		return
	}
	btlLog.BoxChannelInfos.Info("SetTradeChannelsInfoEnd")
}

func CreateLoDataHistory() (err error) {
	return CreateOrUpdateScheduledTasks(&[]models.ScheduledTask{
		{
			Name:           "RecordChannelCount",
			CronExpression: "0 */10 * * * *",
			FunctionName:   "RecordChannelCount",
			Package:        "services",
		},
		{
			Name:           "RecordTotalCapacity",
			CronExpression: "0 */10 * * * *",
			FunctionName:   "RecordTotalCapacity",
			Package:        "services",
		},
		{
			Name:           "RecordNodeCount",
			CronExpression: "0 */10 * * * *",
			FunctionName:   "RecordNodeCount",
			Package:        "services",
		},
		{
			Name:           "UpdateBoxDeviceLocation",
			CronExpression: "1 * * * * *",
			FunctionName:   "UpdateBoxDeviceLocation",
			Package:        "services",
		},
		{
			Name:           "UpdateBoxDeviceLocationEn",
			CronExpression: "31 * * * * *",
			FunctionName:   "UpdateBoxDeviceLocationEn",
			Package:        "services",
		},
		{
			Name:           "UpdateTradeChannelsInfoTime",
			CronExpression: "31 * * * * *",
			FunctionName:   "UpdateTradeChannelsInfoTime",
			Package:        "services",
		},
	})
}

func (cs *CronService) RecordChannelCount() {
	err := lntOfficial.RecordChannelCount()
	if err != nil {
		btlLog.Lnt.Error("RecordChannelCount error: %v", err)
		return
	}
}

func (cs *CronService) RecordTotalCapacity() {
	err := lntOfficial.RecordTotalCapacity()
	if err != nil {
		btlLog.Lnt.Error("RecordTotalCapacity error: %v", err)
		return
	}
}

func (cs *CronService) RecordNodeCount() {
	err := lntOfficial.RecordNodeCount()
	if err != nil {
		btlLog.Lnt.Error("RecordNodeCount error: %v", err)
		return
	}
}

func (cs *CronService) UpdateBoxDeviceLocation() {
	err := lntOfficial.UpdateBoxDeviceLocation()
	if err != nil {
		btlLog.Lnt.Error("UpdateBoxDeviceLocation error: %v", err)
		return
	}
}

func (cs *CronService) UpdateBoxDeviceLocationEn() {
	err := lntOfficial.UpdateBoxDeviceLocationEn()
	if err != nil {
		btlLog.Lnt.Error("UpdateBoxDeviceLocationEn error: %v", err)
		return
	}
}

func (cs *CronService) UpdateTradeChannelsInfoTime() {
	err := lntOfficial.UpdateTradeChannelsInfoTime()
	if err != nil {
		btlLog.Lnt.Error("UpdateTradeChannelsInfoTime error: %v", err)
		return
	}
}
