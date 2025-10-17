package custodyModels

import "gorm.io/gorm"

type AccountBalanceChange struct {
	gorm.Model
	AccountId    uint       `gorm:"column:account_id;type:bigint unsigned;" json:"accountId"`
	AssetId      string     `gorm:"column:asset_id;type:varchar(128);" json:"assetId"`
	ChangeAmount float64    `gorm:"type:decimal(25,2);column:amount" json:"amount"`
	Away         ChangeAway `gorm:"column:away;type:tinyint unsigned" json:"away"`
	FinalBalance float64    `gorm:"type:decimal(25,2);column:final_balance" json:"finalBalance"`
	BalanceId    uint       `gorm:"column:balance_id;type:bigint unsigned;index:idx_balance_id" json:"balanceId"`
	ChangeType   ChangeType `gorm:"column:change_type;type:varchar(128)" json:"changeType"`
}

func (AccountBalanceChange) TableName() string {
	return "user_account_changes"
}

type ChangeAway uint

const (
	ChangeAwayAdd  ChangeAway = 0
	ChangeAwayLess ChangeAway = 1
)

type ChangeType string

const (
	ChangeTypeFault             = "fault"
	ChangeTypeBtcPayOutside     = "pay_outside_btc"
	ChangeTypeBtcReceiveOutside = "receive_outside_btc"
	ChangeTypeBtcPayLocal       = "pay_local_btc"
	ChangeTypeBtcReceiveLocal   = "receive_local_btc"

	ChangeTypeBtcPayOnchain     = "pay_onchain_btc"
	ChangeTypeBtcReceiveOnchain = "receive_onchain_btc"

	ChangeTypeBtcFee        = "btc_fee"
	ChangeTypeAssetFee      = "asset_fee"
	ChangeFirLunchFee       = "fir_lunch_fee"
	ChangeReverseChannelFee = "reverse_channel_fee"

	ChangeTypeBackFee    = "back_fee"
	ChangeTypeAward      = "award"
	ChangeTypeOfferAward = "offer_award"

	ChangePTNSwap         = "ptn_swap"
	ChangePTNSwapSupplier = "ptn_swap_supplier"
	ChangePTNSwapFee      = "ptn_swap_fee"

	ChangeTypePayToPoolAccount       = "pay_to_pool_account"
	ChangeTypeReceiveFromPoolAccount = "receive_from_pool_account"

	ChangeTypeAssetPayOutside     = "pay_outside_asset"
	ChangeTypeAssetPayLocal       = "pay_local_asset"
	ChangeTypeAssetReceiveLocal   = "receive_local_asset"
	ChangeTypeAssetReceiveOutside = "receive_outside_asset"

	ChangeTypeLock           = "lock"
	ChangeTypeUnlock         = "unlock"
	ChangeTypeLockedTransfer = "locked_transfer"

	ChangeTypeReplaceAsset = "replace_asset"

	ClearLimitUser = "clear_limit_user"
)
