package assetMoreInfo

import (
	"strings"
	"time"
	"trade/middleware"
	"trade/models"
	"trade/utils"
)

type BalanceScan struct {
	ID          uint               `json:"id"`
	CreatedAt   time.Time          `json:"created_at"`
	AccountId   uint               `json:"account_id"`
	BillType    models.BalanceType `json:"bill_type"`
	Away        models.BalanceAway `json:"away"`
	Amount      float64            `json:"amount"`
	ServerFee   uint64             `json:"server_fee"`
	AssetId     *string            `json:"asset_id"`
	Invoice     *string            `json:"invoice"`
	PaymentHash *string            `json:"payment_hash"`
}

type AccountAssetTransfer struct {
	BillBalanceId int    `json:"bill_balance_id"`
	AccountId     int    `json:"account_id"`
	Username      string `json:"username"`
	BillType      string `json:"bill_type"`
	Away          string `json:"away"`
	Amount        int    `json:"amount"`
	ServerFee     int    `json:"server_fee"`
	AssetId       string `json:"asset_id"`
	Invoice       string `json:"invoice"`
	Outpoint      string `json:"outpoint"`
	Time          int    `json:"time"`
}

func BalanceScanToAccountAssetTransfer(balanceScan BalanceScan, username string) AccountAssetTransfer {
	if balanceScan.ID == 0 {
		return AccountAssetTransfer{}
	}
	var assetId string
	if balanceScan.AssetId != nil {
		assetId = *balanceScan.AssetId
	}
	var invoice string
	if balanceScan.Invoice != nil {
		invoice = *balanceScan.Invoice
	}
	var outpoint string
	if balanceScan.PaymentHash != nil && *balanceScan.PaymentHash != "" && strings.Contains(*balanceScan.PaymentHash, ":") {
		outpoint = *balanceScan.PaymentHash
	}
	return AccountAssetTransfer{
		BillBalanceId: int(balanceScan.ID),
		AccountId:     int(balanceScan.AccountId),
		Username:      username,
		BillType:      balanceScan.BillType.String(),
		Away:          balanceScan.Away.String(),
		Amount:        int(balanceScan.Amount),
		ServerFee:     int(balanceScan.ServerFee),
		AssetId:       assetId,
		Invoice:       invoice,
		Outpoint:      outpoint,
		Time:          int(balanceScan.CreatedAt.Unix()),
	}
}

func GetBalanceScans(assetId string, limit int, offset int) (balanceScans []BalanceScan, err error) {
	err = middleware.DB.
		Table("bill_balance").
		Select("id, created_at, account_id, bill_type, away, amount, server_fee, asset_id, invoice, payment_hash").
		Where("amount <> ? and bill_type in ? and asset_id = ?", 0, []models.BalanceType{models.BillTypeAssetTransfer, models.BillTypeAwardAsset}, assetId).
		Order("updated_at desc").
		Limit(limit).
		Offset(offset).
		Scan(&balanceScans).
		Error
	return balanceScans, err
}

type UserIdAndUsername struct {
	UserId   int    `json:"user_id"`
	Username string `json:"username"`
}

func ReadUserAccountAccountId(accountId uint) (*models.Account, error) {
	var account models.Account
	err := middleware.DB.First(&account, accountId).Error
	return &account, err
}

func GetUserIdAndUsernameByAccountId(accountId uint) (*UserIdAndUsername, error) {
	account, err := ReadUserAccountAccountId(accountId)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "ReadUserAccountAccountId")
	}
	return &UserIdAndUsername{
		UserId:   int(account.UserId),
		Username: account.UserName,
	}, nil
}

func BalanceScansToAccountAssetTransfers(balanceScans []BalanceScan) []AccountAssetTransfer {
	if len(balanceScans) == 0 {
		return nil
	}
	var accountAssetTransfers []AccountAssetTransfer
	for _, balanceScan := range balanceScans {
		var usernameByAccountId string
		userIdAndUsername, err := GetUserIdAndUsernameByAccountId(balanceScan.AccountId)
		if err != nil {
			continue
		} else {
			usernameByAccountId = userIdAndUsername.Username
		}
		accountAssetTransfers = append(accountAssetTransfers, BalanceScanToAccountAssetTransfer(balanceScan, usernameByAccountId))
	}
	return accountAssetTransfers
}

func GetAccountAssetTransferCount(assetId string) (count int64, err error) {
	err = middleware.DB.
		Table("bill_balance").
		Where("amount <> ? and asset_id = ?", 0, assetId).
		Count(&count).
		Error
	return count, err
}

func GetAccountAssetTransfer(assetId string, limit int, offset int) (accountAssetTransfers []AccountAssetTransfer, err error) {
	balanceScans, err := GetBalanceScans(assetId, limit, offset)
	if err != nil {
		return nil, err
	}
	accountAssetTransfers = BalanceScansToAccountAssetTransfers(balanceScans)
	return accountAssetTransfers, nil
}
