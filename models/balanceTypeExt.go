package models

type BalanceTypeExt struct {
	BalanceID uint               `json:"balance" gorm:"not null;"`
	Type      BalanceTypeExtList `json:"type" gorm:"not null"`
}

func (BalanceTypeExt) TableName() string {
	return "bill_balance_type_ext"
}

type BalanceTypeExtList uint

const (
	BTExtUnknown            BalanceTypeExtList = 0
	BTExtFirLaunch          BalanceTypeExtList = 6
	BTExtLocal              BalanceTypeExtList = 100
	BTExtBackFee            BalanceTypeExtList = 104
	BTExtOnChannel          BalanceTypeExtList = 200
	BTExtAward              BalanceTypeExtList = 300
	BTExtLocked             BalanceTypeExtList = 400
	BTExtLockedTransfer     BalanceTypeExtList = 500
	BTExtPayToPoolAccount   BalanceTypeExtList = 600
	BTExtReceivePoolAccount BalanceTypeExtList = 601
	BTEServerFee            BalanceTypeExtList = 700
	BTEFirLunchFee          BalanceTypeExtList = 701
	BTEFirBackFee           BalanceTypeExtList = 702
)

func (b BalanceTypeExtList) ToString() string {
	balanceTypeExtString := map[BalanceTypeExtList]string{
		BTExtUnknown:            "Unknown",
		BTExtFirLaunch:          "FirLaunch",
		BTExtLocal:              "Local",
		BTExtBackFee:            "BackFee",
		BTExtOnChannel:          "OnChannel",
		BTExtAward:              "Award",
		BTExtLocked:             "Locked",
		BTExtLockedTransfer:     "LockedTransfer",
		BTExtPayToPoolAccount:   "PayToPoolAccount",
		BTExtReceivePoolAccount: "ReceivePoolAccount",
		BTEServerFee:            "ServerFee",
		BTEFirLunchFee:          "FirLunchFee",
		BTEFirBackFee:           "FirBackFee",
	}
	return balanceTypeExtString[b]
}
func ToBalanceTypeExtList(s string) BalanceTypeExtList {
	balanceTypeExtList := map[string]BalanceTypeExtList{
		"Unknown":            BTExtUnknown,
		"FirLaunch":          BTExtFirLaunch,
		"Local":              BTExtLocal,
		"BackFee":            BTExtBackFee,
		"OnChannel":          BTExtOnChannel,
		"Award":              BTExtAward,
		"Locked":             BTExtLocked,
		"LockedTransfer":     BTExtLockedTransfer,
		"PayToPoolAccount":   BTExtPayToPoolAccount,
		"ReceivePoolAccount": BTExtReceivePoolAccount,
		"ServerFee":          BTEServerFee,
		"FirLunchFee":        BTEFirLunchFee,
		"FirBackFee":         BTEFirBackFee,
	}
	return balanceTypeExtList[s]
}
