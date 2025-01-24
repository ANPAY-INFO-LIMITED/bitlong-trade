package custodyBtc

import (
	"fmt"
	"github.com/lightningnetwork/lnd/lnrpc"
	"gorm.io/gorm"
	"time"
	"trade/btlLog"
	"trade/config"
	"trade/middleware"
	"trade/models"
	"trade/models/custodyModels"
	"trade/services/custodyAccount/account"
	"trade/services/custodyAccount/custodyBase/custodyFee"
	"trade/services/custodyAccount/custodyBase/custodyLimit"
	"trade/services/custodyAccount/custodyBase/custodyRpc"
	"trade/services/custodyAccount/defaultAccount/custodyBtc/mempool"
	"trade/services/servicesrpc"
)

func RunOutsideSteps(usr *account.UserInfo, mission *custodyModels.AccountOutsideMission) error {
	db := middleware.DB
	if usr == nil {
		var a models.Account
		if err := db.Where("id =?", mission.AccountId).First(&a).Error; err != nil {
			btlLog.CUST.Error("GetAccount error:%s", err)
			return err
		}
		usr, _ = account.GetUserInfo(a.UserName)
	}
	for {
		OutsideSteps(usr, mission)
		LogAOM(db, mission)
		switch {
		case mission.State == custodyModels.AOMStateSuccess:
			db.Model(&models.Balance{}).
				Where("id = ?", mission.BalanceId). // 根据需要的条件
				Updates(models.Balance{ServerFee: uint64(mission.Fee), State: models.STATE_SUCCESS})

			limitType := custodyModels.LimitType{
				AssetId:      "00",
				TransferType: custodyModels.LimitTransferTypeOutside,
			}
			_ = custodyLimit.MinusLimit(db, usr, &limitType, mission.Amount+mission.Fee)
			return nil
		case mission.State == custodyModels.AOMStateDone:
			db.Table("bill_balance").Where("id =?", mission.BalanceId).Update("State", models.STATE_FAILED)
			return fmt.Errorf(mission.Error)
		case mission.Retries >= 30:
			return nil
		}
	}
}

func OutsideSteps(usr *account.UserInfo, mission *custodyModels.AccountOutsideMission) {
	var err error
	switch mission.State {
	case custodyModels.AOMStatePending:
		tx, back := middleware.GetTx()
		defer back()
		_, err = LessBtcBalance(tx, usr, mission.Amount, mission.BalanceId, custodyModels.ChangeTypeBtcPayOutside)
		if err != nil {
			btlLog.CUST.Error("PayBtcInvoice error:%s", err)
			mission.State = custodyModels.AOMStateDone
			mission.Error = err.Error()
			return
		}
		payment, err := custodyRpc.PayBtcInvoice(usr, mission.Target, int64(mission.Amount), int64(mission.FeeLimit))
		if err != nil || payment.Status != lnrpc.Payment_SUCCEEDED {
			if err == nil {
				err = fmt.Errorf("PayBtcInvoice error:%s", payment.FailureReason.String())
			}
			btlLog.CUST.Error("PayBtcInvoice error:%s", err)
			mission.State = custodyModels.AOMStateDone
			mission.Error = err.Error()
			return
		}
		tx.Commit()
		go subscriptionLndBalance(float64(payment.ValueSat + payment.FeeSat))

		//todo 检查如果发票已被使用，则需要查询发票以获取相关信息
		mission.Fee = float64(payment.FeeSat)
		mission.FeeType = "00"
		mission.State = custodyModels.AOMStateNotPayFee
		return

	case custodyModels.AOMStateNotPayFee:
		db := middleware.DB
		mission.Fee += float64(custodyFee.ChannelBtcServiceFee)
		err := PayFee(db, usr, mission.Fee, mission.BalanceId, &mission.Target, &mission.Hash)
		if err != nil {
			btlLog.CUST.Error("PayBtcFeeError:%s", err)
			mission.Retries += 1
			return
		}
		mission.State = custodyModels.AOMStateSuccess
		return
	}
}

func LoadAOMMission() {
	var missions []custodyModels.AccountOutsideMission
	middleware.DB.Where("state =? OR state =?", custodyModels.AOMStatePending, custodyModels.AOMStateNotPayFee).Find(&missions)
	for _, m := range missions {
		_ = RunOutsideSteps(nil, &m)
	}
}

func LogAOM(tx *gorm.DB, mission *custodyModels.AccountOutsideMission) {
	tx.Save(mission)
}

func subscriptionLndBalance(amount float64) {
	if config.GetLoadConfig().NetWork == "regtest" {
		return
	}
	time.Sleep(time.Second * 10)
	d := mempool.NewDingding()
	var balances []mempool.Balance

	channels, err := servicesrpc.GetChannelInfo()
	if err != nil {
		btlLog.CUST.Error("GetChannelInfo error:%s", err)
		return
	}
	for _, c := range channels {
		if c.LocalBalance >= 0 {
			balances = append(balances, mempool.Balance{
				Name:  c.PeerAlias,
				Value: float64(c.LocalBalance),
			})
		}
	}

	balance, err := servicesrpc.GetBalance()
	if err != nil {
		btlLog.CUST.Error("GetChannelInfo error:%s", err)
		return
	}
	if balance != nil && len(balance.AccountBalance) > 0 {
		balances = append(balances, mempool.Balance{
			Name:  "链上余额",
			Value: float64(balance.AccountBalance["default"].ConfirmedBalance),
		})
		balances = append(balances, mempool.Balance{
			Name:  "链上未确认余额",
			Value: float64(balance.AccountBalance["default"].UnconfirmedBalance),
		})
	}

	abalance, err := servicesrpc.ListAssetsBalance()
	if err != nil {
		btlLog.CUST.Error("ListAssetsBalance error:%s", err)
		return
	}
	if abalance != nil && len(abalance.AssetBalances) > 0 {
		for _, b := range abalance.AssetBalances {
			if b.AssetGenesis.Name == "Phenix" {
				balances = append(balances, mempool.Balance{
					Name:  "Phenix",
					Value: float64(b.Balance),
				})
				break
			}
		}
	}

	err = d.SendBtcPayOutChange(amount, balances)
	if err != nil {
		btlLog.CUST.Error("SendBtcPayOutChange error:%s", err)
	}
}
